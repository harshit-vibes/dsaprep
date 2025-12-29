package cfweb

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	v1 "github.com/harshit-vibes/cf/pkg/internal/schema/v1"
)

// Pre-compiled regexes for performance
var (
	reTitlePrefix    = regexp.MustCompile(`^[A-Z]\d*\.\s*`)
	reWhitespace     = regexp.MustCompile(`\s+`)
	reProblemIndex   = regexp.MustCompile(`/problem/([A-Z]\d*)$`)
	reRating         = regexp.MustCompile(`\*(\d+)`)
)

// Parser scrapes problem data from CF web pages
type Parser struct {
	session   *Session
	selectors Selectors
}

// NewParser creates a new parser
func NewParser(session *Session) *Parser {
	return &Parser{
		session:   session,
		selectors: CurrentSelectors,
	}
}

// NewParserWithClient creates a parser with a custom HTTP client
func NewParserWithClient(client *http.Client) *Parser {
	return &Parser{
		session: &Session{
			client: client,
		},
		selectors: CurrentSelectors,
	}
}

// ParsedProblem contains parsed problem data
type ParsedProblem struct {
	ContestID   int
	Index       string
	Name        string
	TimeLimit   string
	MemoryLimit string
	Statement   string
	InputSpec   string
	OutputSpec  string
	Note        string
	Samples     []Sample
	Tags        []string
	Rating      int
	URL         string
}

// Sample represents a test case
type Sample struct {
	Index  int
	Input  string
	Output string
}

// ParseProblem parses a problem page
func (p *Parser) ParseProblem(contestID int, index string) (*ParsedProblem, error) {
	// Construct problem URL
	url := fmt.Sprintf("%s/contest/%d/problem/%s", BaseURL, contestID, index)

	resp, err := p.fetch(url)
	if err != nil {
		return nil, fmt.Errorf("fetch problem page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("problem page returned status %d", resp.StatusCode)
	}

	return p.parseProblemHTML(resp.Body, contestID, index, url)
}

// ParseProblemset parses a problem from the problemset
func (p *Parser) ParseProblemset(contestID int, index string) (*ParsedProblem, error) {
	url := fmt.Sprintf("%s/problemset/problem/%d/%s", BaseURL, contestID, index)

	resp, err := p.fetch(url)
	if err != nil {
		return nil, fmt.Errorf("fetch problemset page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("problemset page returned status %d", resp.StatusCode)
	}

	return p.parseProblemHTML(resp.Body, contestID, index, url)
}

// parseProblemHTML parses the problem HTML
func (p *Parser) parseProblemHTML(r io.Reader, contestID int, index, url string) (*ParsedProblem, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	sel := p.selectors.Problem

	problem := &ParsedProblem{
		ContestID: contestID,
		Index:     index,
		URL:       url,
	}

	// Parse title
	titleText := doc.Find(sel.Title).First().Text()
	problem.Name = cleanTitle(titleText)

	// Parse time limit
	timeLimitText := doc.Find(sel.TimeLimit).First().Text()
	problem.TimeLimit = extractLimit(timeLimitText, "time limit per test")

	// Parse memory limit
	memoryLimitText := doc.Find(sel.MemoryLimit).First().Text()
	problem.MemoryLimit = extractLimit(memoryLimitText, "memory limit per test")

	// Parse statement - get the full problem statement div
	statementNode := doc.Find(sel.Statement).First()

	// Get input specification
	inputSpec := doc.Find(sel.InputSpec).First()
	problem.InputSpec = cleanHTML(inputSpec.Text())

	// Get output specification
	outputSpec := doc.Find(sel.OutputSpec).First()
	problem.OutputSpec = cleanHTML(outputSpec.Text())

	// Get note
	note := doc.Find(sel.Note).First()
	problem.Note = cleanHTML(note.Text())

	// Build statement from parts
	problem.Statement = buildStatement(statementNode)

	// Parse samples
	sampleTests := doc.Find(sel.SampleTests)
	if sampleTests.Length() > 0 {
		problem.Samples = parseSamples(sampleTests, sel)
	}

	// Parse tags
	doc.Find(sel.Tags).Each(func(i int, s *goquery.Selection) {
		tag := strings.TrimSpace(s.Text())
		// Skip rating tag and empty tags
		if tag != "" && !strings.HasPrefix(tag, "*") {
			problem.Tags = append(problem.Tags, tag)
		}
	})

	// Parse rating
	ratingText := ""
	doc.Find(sel.Rating).Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.HasPrefix(text, "*") {
			ratingText = text
		}
	})
	if ratingText != "" {
		problem.Rating = parseRating(ratingText)
	}

	return problem, nil
}

// ParseContestProblems parses all problems from a contest
func (p *Parser) ParseContestProblems(contestID int) ([]ParsedProblem, error) {
	url := fmt.Sprintf("%s/contest/%d", BaseURL, contestID)

	resp, err := p.fetch(url)
	if err != nil {
		return nil, fmt.Errorf("fetch contest page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("contest page returned status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	var problems []ParsedProblem
	sel := p.selectors.Contest

	doc.Find(sel.ProblemRow).Each(func(i int, s *goquery.Selection) {
		// Skip header row
		if i == 0 {
			return
		}

		linkSel := s.Find(sel.ProblemLink).First()
		href, exists := linkSel.Attr("href")
		if !exists {
			return
		}

		// Extract problem index from href
		index := extractProblemIndex(href)
		if index == "" {
			return
		}

		// Extract problem name
		nameSel := s.Find("td").Eq(1).Find("a").First()
		name := strings.TrimSpace(nameSel.Text())

		problems = append(problems, ParsedProblem{
			ContestID: contestID,
			Index:     index,
			Name:      name,
			URL:       BaseURL + href,
		})
	})

	return problems, nil
}

// fetch makes an HTTP GET request
func (p *Parser) fetch(url string) (*http.Response, error) {
	if p.session != nil && p.session.client != nil {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", UserAgent)
		return p.session.client.Do(req)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	return http.DefaultClient.Do(req)
}

// ToSchemaProblem converts ParsedProblem to schema v1 Problem
func (p *ParsedProblem) ToSchemaProblem() *v1.Problem {
	samples := make([]v1.Sample, len(p.Samples))
	for i, s := range p.Samples {
		samples[i] = v1.Sample{
			Index:  s.Index,
			Input:  s.Input,
			Output: s.Output,
		}
	}

	return &v1.Problem{
		ID:        fmt.Sprintf("%d%s", p.ContestID, p.Index),
		Platform:  "codeforces",
		ContestID: p.ContestID,
		Index:     p.Index,
		Name:      p.Name,
		URL:       p.URL,
		Limits: v1.ProblemLimits{
			TimeLimit:   p.TimeLimit,
			MemoryLimit: p.MemoryLimit,
		},
		Metadata: v1.ProblemMetadata{
			Rating: p.Rating,
			Tags:   p.Tags,
		},
		Samples: samples,
		// Note: Statement is saved separately as statement.md
	}
}

// Helper functions

func cleanTitle(title string) string {
	// Remove problem index prefix like "A. " or "B1. "
	title = reTitlePrefix.ReplaceAllString(strings.TrimSpace(title), "")
	return strings.TrimSpace(title)
}

func extractLimit(text, prefix string) string {
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, prefix)
	return strings.TrimSpace(text)
}

func cleanHTML(text string) string {
	// Remove extra whitespace
	text = strings.TrimSpace(text)
	// Collapse multiple spaces
	text = reWhitespace.ReplaceAllString(text, " ")
	return text
}

func buildStatement(statement *goquery.Selection) string {
	if statement == nil {
		return ""
	}

	var sb strings.Builder

	// Get direct text content, excluding child elements we handle separately
	statement.Contents().Each(func(i int, s *goquery.Selection) {
		if goquery.NodeName(s) == "#text" {
			text := strings.TrimSpace(s.Text())
			if text != "" {
				sb.WriteString(text)
				sb.WriteString(" ")
			}
		}
	})

	return strings.TrimSpace(sb.String())
}

func parseSamples(sampleTests *goquery.Selection, sel ProblemSelectors) []Sample {
	var samples []Sample
	sampleIdx := 1

	// Find all input-output pairs
	sampleTests.Find(".sample-test").Each(func(i int, s *goquery.Selection) {
		// Each sample-test contains input and output divs
		inputDiv := s.Find(".input pre")
		outputDiv := s.Find(".output pre")

		if inputDiv.Length() > 0 && outputDiv.Length() > 0 {
			sample := Sample{
				Index:  sampleIdx,
				Input:  extractPreContent(inputDiv),
				Output: extractPreContent(outputDiv),
			}
			samples = append(samples, sample)
			sampleIdx++
		}
	})

	// Alternative structure: direct input/output pairs
	if len(samples) == 0 {
		inputs := sampleTests.Find(".input pre")
		outputs := sampleTests.Find(".output pre")

		minLen := inputs.Length()
		if outputs.Length() < minLen {
			minLen = outputs.Length()
		}

		for i := 0; i < minLen; i++ {
			sample := Sample{
				Index:  i + 1,
				Input:  extractPreContent(inputs.Eq(i)),
				Output: extractPreContent(outputs.Eq(i)),
			}
			samples = append(samples, sample)
		}
	}

	return samples
}

func extractPreContent(sel *goquery.Selection) string {
	// CF sometimes uses <br> tags instead of newlines
	html, _ := sel.Html()

	// Replace <br> with newlines
	html = strings.ReplaceAll(html, "<br/>", "\n")
	html = strings.ReplaceAll(html, "<br>", "\n")

	// Parse the HTML and extract text
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return strings.TrimSpace(sel.Text())
	}

	text := doc.Text()

	// Clean up the text
	lines := strings.Split(text, "\n")
	var cleaned []string
	for _, line := range lines {
		// Don't trim leading spaces as they might be significant
		line = strings.TrimRight(line, " \t")
		cleaned = append(cleaned, line)
	}

	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

func extractProblemIndex(href string) string {
	// Extract problem index from URLs like /contest/123/problem/A
	matches := reProblemIndex.FindStringSubmatch(href)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func parseRating(text string) int {
	// Extract rating from text like "*1400"
	matches := reRating.FindStringSubmatch(text)
	if len(matches) > 1 {
		rating, _ := strconv.Atoi(matches[1])
		return rating
	}
	return 0
}

// VerifyPageStructure checks if the page structure matches expected selectors
func (p *Parser) VerifyPageStructure() error {
	// Test with a known problem
	url := fmt.Sprintf("%s/problemset/problem/1/A", BaseURL)

	resp, err := p.fetch(url)
	if err != nil {
		return fmt.Errorf("fetch test page: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("parse HTML: %w", err)
	}

	sel := p.selectors.Problem

	// Verify critical selectors
	checks := map[string]string{
		"title":        sel.Title,
		"time_limit":   sel.TimeLimit,
		"memory_limit": sel.MemoryLimit,
		"statement":    sel.Statement,
		"samples":      sel.SampleTests,
	}

	var missing []string
	for name, selector := range checks {
		if doc.Find(selector).Length() == 0 {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("selectors not found: %s (selector version: %s)",
			strings.Join(missing, ", "), CurrentVersion.Version)
	}

	return nil
}
