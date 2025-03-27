package server

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/cnosuke/mcp-fetch/config"
	"github.com/cnosuke/mcp-fetch/types"
	"github.com/cockroachdb/errors"
	"github.com/go-shiori/go-readability"
	"go.uber.org/zap"
)

// FetchServer - Fetch server structure
type FetchServer struct {
	client     *http.Client
	userAgent  string
	maxWorkers int
}

// NewFetchServer - Create a new Fetch server
func NewFetchServer(cfg *config.Config) (*FetchServer, error) {
	zap.S().Infow("creating new Fetch server",
		"timeout", cfg.Fetch.Timeout,
		"user_agent", cfg.Fetch.UserAgent,
		"max_workers", cfg.Fetch.MaxWorkers)

	client := &http.Client{
		Timeout: time.Duration(cfg.Fetch.Timeout) * time.Second,
	}

	return &FetchServer{
		client:     client,
		userAgent:  cfg.Fetch.UserAgent,
		maxWorkers: cfg.Fetch.MaxWorkers,
	}, nil
}

// FetchURL - Fetch content from a URL with content control options
func (s *FetchServer) FetchURL(url string, maxLength int, startIndex int, raw bool) (*types.FetchResponse, error) {
	zap.S().Debugw("fetching URL",
		"url", url,
		"max_length", maxLength,
		"start_index", startIndex,
		"raw", raw)

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	// Set headers
	req.Header.Set("User-Agent", s.userAgent)

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute request")
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	zap.S().Debugw(
		"response received",
		"url", url,
		"status", resp.StatusCode,
		"content-length", resp.ContentLength,
		"bytes", len(body),
		"content_type", resp.Header.Get("Content-Type"),
	)

	contentType := resp.Header.Get("Content-Type")
	content := string(body)

	// Process content based on content type and raw flag
	if !raw && strings.Contains(contentType, "text/html") {
		processedContent, err := s.processHTMLContent(content, url)
		if err != nil {
			zap.S().Warnw("failed to process HTML content with readability, falling back to basic conversion", "error", err)
			// If readability fails, try fallback HTML-to-markdown conversion
			basicMarkdown, fallbackErr := s.convertHTMLToMarkdown(content)
			if fallbackErr != nil {
				zap.S().Warnw("fallback HTML conversion also failed", "error", fallbackErr)
				// If all conversions fail, return the original content
			} else {
				content = basicMarkdown
			}
		} else {
			content = processedContent
		}
	} else if !raw {
		// For non-HTML content, just return as is
		content = s.processNonHTMLContent(content, contentType)
	} else {
		// Raw mode - return content as-is
		zap.S().Debugw("raw mode enabled, returning content as-is",
			"url", url,
			"content_type", contentType)
	}

	// Apply content trimming based on startIndex and maxLength
	if len(content) > 0 {
		// Validate startIndex
		if startIndex < 0 {
			startIndex = 0
		}
		if startIndex > len(content) {
			startIndex = len(content)
		}

		// Apply maxLength if specified
		endIndex := len(content)
		if maxLength > 0 && startIndex+maxLength < endIndex {
			endIndex = startIndex + maxLength
		}

		// Trim content
		if startIndex > 0 || endIndex < len(content) {
			originalLength := len(content)
			content = content[startIndex:endIndex]
			zap.S().Debugw("content trimmed",
				"original_length", originalLength,
				"start_index", startIndex,
				"end_index", endIndex,
				"trimmed_length", len(content))
		}
	}

	return &types.FetchResponse{
		URL:         url,
		ContentType: contentType,
		Content:     content,
		StatusCode:  resp.StatusCode,
	}, nil
}

// processHTMLContent - Process HTML content using readability and convert to markdown
func (s *FetchServer) processHTMLContent(htmlContent, urlStr string) (string, error) {
	// Parse URL string to *url.URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse URL")
	}

	// Use readability to extract the main content
	article, err := readability.FromReader(strings.NewReader(htmlContent), parsedURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to extract content with readability")
	}

	// Convert the extracted content to Markdown
	converter := md.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(article.Content)
	if err != nil {
		return article.Content, errors.Wrap(err, "failed to convert extracted content to Markdown")
	}

	// Add title as heading if available
	if article.Title != "" {
		markdown = "# " + article.Title + "\n\n" + markdown
	}

	// Add excerpt/description if available
	if article.Excerpt != "" {
		markdown = markdown + "\n\n---\n\n" + article.Excerpt
	}

	zap.S().Debugw(
		"processed HTML content",
		"url", urlStr,
		"title", article.Title,
		"length", len(markdown),
	)

	return markdown, nil
}

// processNonHTMLContent - Process non-HTML content
func (s *FetchServer) processNonHTMLContent(content, contentType string) string {
	// For now, we just return the content as is
	// But we could add more processing for different content types in the future
	zap.S().Debugw(
		"non-HTML content detected, returning as is",
		"content_type", contentType,
		"length", len(content),
	)
	return content
}

// convertHTMLToMarkdown - Convert HTML to Markdown directly
func (s *FetchServer) convertHTMLToMarkdown(htmlContent string) (string, error) {
	converter := md.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(htmlContent)
	if err != nil {
		return htmlContent, errors.Wrap(err, "failed to convert HTML to Markdown")
	}
	return markdown, nil
}

// URLFetchResult - Struct to store fetching results with allocation information
type URLFetchResult struct {
	URL            string
	Response       *types.FetchResponse
	Error          error
	AllocatedChars int // 割り当てられた文字数
	UsedChars      int // 実際に使用した文字数
}

// FetchMultipleURLs - Fetch content from multiple URLs in parallel with content control and reallocation
func (s *FetchServer) FetchMultipleURLs(urls []string, maxLength int, raw bool) (*types.MultipleFetchResponse, error) {
	zap.S().Debugw("fetching multiple URLs with reallocation",
		"count", len(urls),
		"max_length", maxLength,
		"raw", raw,
		"workers", s.maxWorkers)

	// maxLengthが指定されていない場合のデフォルト値
	if maxLength <= 0 {
		maxLength = 5000
	}

	// 各URLの初期配分文字数の計算
	var initialAllocation int
	if len(urls) > 0 {
		initialAllocation = maxLength / len(urls)
	}

	zap.S().Debugw("initial allocation per URL calculated",
		"initial_allocation", initialAllocation,
		"total_max_length", maxLength)

	// 結果を格納するためのスライス
	results := make([]*URLFetchResult, len(urls))
	for i := range results {
		results[i] = &URLFetchResult{
			URL:            urls[i],
			AllocatedChars: initialAllocation,
		}
	}

	// 待機グループ
	wg := &sync.WaitGroup{}

	// 並列処理によるURLの取得
	for i, result := range results {
		wg.Add(1)
		go func(index int, r *URLFetchResult) {
			defer wg.Done()

			zap.S().Debugw("fetching URL with initial allocation",
				"url", r.URL,
				"initial_allocation", r.AllocatedChars)

			// 初期配分量で取得
			response, err := s.FetchURL(r.URL, r.AllocatedChars, 0, raw)

			r.Response = response
			r.Error = err

			if err == nil && response != nil {
				r.UsedChars = len(response.Content)

				// 割り当てられた量より少ない場合はログを出力
				if r.UsedChars < r.AllocatedChars {
					zap.S().Debugw("URL used less than allocated length",
						"url", r.URL,
						"allocated", r.AllocatedChars,
						"used", r.UsedChars,
						"saved", r.AllocatedChars-r.UsedChars)
				}
			}
		}(i, result)
	}

	// すべてのURLの処理が完了するまで待機
	wg.Wait()

	// 再配分のための未使用文字数の計算
	var totalUsed int
	var redistribution []*URLFetchResult // 再配分が可能なURL
	var beneficiaries []*URLFetchResult  // 再配分を受け取れるURL

	for _, r := range results {
		if r.Error != nil {
			// エラーがあった場合は、割り当て分はすべて未使用とする
			continue
		}

		totalUsed += r.UsedChars

		// 未使用分がある場合は再配分候補に追加
		if r.UsedChars < r.AllocatedChars {
			redistribution = append(redistribution, r)
		} else if r.Response != nil {
			// まだコンテンツが切り詰められている可能性があるURLは受益者に追加
			// (レスポンスの内容が切り詰められていない限り)
			beneficiaries = append(beneficiaries, r)
		}
	}

	// 再配分可能な総文字数
	remainingChars := maxLength - totalUsed

	zap.S().Debugw("redistribution status after initial fetch",
		"total_used", totalUsed,
		"total_max", maxLength,
		"remaining_chars", remainingChars,
		"redistribution_urls", len(redistribution),
		"beneficiary_urls", len(beneficiaries))

	// 再配分が可能で、受益者が存在する場合に再配分を実行
	if remainingChars > 0 && len(beneficiaries) > 0 {
		// 受益者を並べ替えない場合は、単純に均等に配分
		perURLReallocation := remainingChars / len(beneficiaries)

		if perURLReallocation > 0 {
			zap.S().Debugw("performing redistribution",
				"per_url_reallocation", perURLReallocation,
				"beneficiary_count", len(beneficiaries))

			// 各受益者に再配分
			for _, b := range beneficiaries {
				// 追加分を取得
				additionalContent, err := s.fetchAdditionalContent(b.URL, b.Response, perURLReallocation, raw)
				if err != nil {
					zap.S().Warnw("failed to fetch additional content",
						"url", b.URL,
						"error", err)
					continue
				}

				// 既存のコンテンツに追加分を追加
				originalLength := len(b.Response.Content)
				b.Response.Content += additionalContent
				b.UsedChars = len(b.Response.Content)

				zap.S().Debugw("additional content fetched and appended",
					"url", b.URL,
					"original_length", originalLength,
					"additional_length", len(additionalContent),
					"new_total_length", b.UsedChars)
			}
		}
	}

	// 最終的な結果を構築
	response := &types.MultipleFetchResponse{
		Responses: make(map[string]*types.FetchResponse),
		Errors:    make(map[string]string),
	}

	// 結果を格納
	for _, r := range results {
		if r.Error != nil {
			response.Errors[r.URL] = r.Error.Error()
		} else if r.Response != nil {
			response.Responses[r.URL] = r.Response
		}
	}

	// 完了ログ
	zap.S().Infow("completed fetching multiple URLs with reallocation",
		"total_urls", len(urls),
		"success", len(response.Responses),
		"errors", len(response.Errors),
		"total_content_length", getTotalContentLength(response))

	return response, nil
}

// fetchAdditionalContent - 追加のコンテンツを取得
func (s *FetchServer) fetchAdditionalContent(url string, originalResponse *types.FetchResponse, additionalChars int, raw bool) (string, error) {
	if additionalChars <= 0 || originalResponse == nil {
		return "", nil
	}

	// 続きから取得するためにオフセットを設定
	startIndex := len(originalResponse.Content)

	// 追加分を取得
	response, err := s.FetchURL(url, additionalChars, startIndex, raw)
	if err != nil {
		return "", err
	}

	return response.Content, nil
}

// getTotalContentLength - レスポンス内の全コンテンツの合計長を取得
func getTotalContentLength(response *types.MultipleFetchResponse) int {
	if response == nil {
		return 0
	}

	var total int
	for _, resp := range response.Responses {
		total += len(resp.Content)
	}

	return total
}
