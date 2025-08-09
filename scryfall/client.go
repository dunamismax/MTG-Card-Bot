// Package scryfall provides a client for interacting with the Scryfall API.
package scryfall

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dunamismax/MTG-Card-Bot/errors"
	"github.com/dunamismax/MTG-Card-Bot/logging"
	"github.com/dunamismax/MTG-Card-Bot/metrics"
)

const (
	// BaseURL is the base URL for the Scryfall API.
	BaseURL = "https://api.scryfall.com"
	// UserAgent is the User-Agent header for API requests.
	UserAgent = "MTGDiscordBot/1.0"
	// RateLimit defines the rate limit for API requests (10 requests per second as recommended).
	RateLimit = 100 * time.Millisecond
)

// Client represents a Scryfall API client with rate limiting.
type Client struct {
	httpClient  *http.Client
	rateLimiter *time.Ticker
}

// Card represents a Magic: The Gathering card from the Scryfall API.
type Card struct {
	Object       string            `json:"object"`
	ID           string            `json:"id"`
	OracleID     string            `json:"oracle_id"`
	Name         string            `json:"name"`
	Lang         string            `json:"lang"`
	ReleasedAt   string            `json:"released_at"`
	URI          string            `json:"uri"`
	ScryfallURI  string            `json:"scryfall_uri"`
	Layout       string            `json:"layout"`
	ImageUris    map[string]string `json:"image_uris,omitempty"`
	CardFaces    []CardFace        `json:"card_faces,omitempty"`
	ManaCost     string            `json:"mana_cost,omitempty"`
	CMC          float64           `json:"cmc"`
	TypeLine     string            `json:"type_line"`
	OracleText   string            `json:"oracle_text,omitempty"`
	Colors       []string          `json:"colors,omitempty"`
	SetName      string            `json:"set_name"`
	SetCode      string            `json:"set"`
	Rarity       string            `json:"rarity"`
	Artist       string            `json:"artist,omitempty"`
	Prices       Prices            `json:"prices"`
	ImageStatus  string            `json:"image_status"`
	HighresImage bool              `json:"highres_image"`
}

// CardFace represents one face of a multi-faced card.
type CardFace struct {
	Object     string            `json:"object"`
	Name       string            `json:"name"`
	ManaCost   string            `json:"mana_cost"`
	TypeLine   string            `json:"type_line"`
	OracleText string            `json:"oracle_text,omitempty"`
	Colors     []string          `json:"colors,omitempty"`
	Artist     string            `json:"artist,omitempty"`
	ImageUris  map[string]string `json:"image_uris,omitempty"`
}

// Prices represents the pricing information for a card.
type Prices struct {
	USD     *string `json:"usd"`
	USDFoil *string `json:"usd_foil"`
	EUR     *string `json:"eur"`
	EURFoil *string `json:"eur_foil"`
	Tix     *string `json:"tix"`
}

// SearchResult represents the result of a card search query.
type SearchResult struct {
	Object     string `json:"object"`
	TotalCards int    `json:"total_cards"`
	HasMore    bool   `json:"has_more"`
	NextPage   string `json:"next_page,omitempty"`
	Data       []Card `json:"data"`
}

// Error represents an error response from the Scryfall API.
type Error struct {
	Object   string   `json:"object"`
	Code     string   `json:"code"`
	Status   int      `json:"status"`
	Details  string   `json:"details"`
	Type     string   `json:"type,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

func (e Error) Error() string {
	return fmt.Sprintf("scryfall api error: %s (status: %d)", e.Details, e.Status)
}

// GetErrorType returns the error type for metrics tracking.
func (e Error) GetErrorType() errors.ErrorType {
	switch e.Status {
	case 404:
		return errors.ErrorTypeNotFound
	case 429:
		return errors.ErrorTypeRateLimit
	default:
		return errors.ErrorTypeAPI
	}
}

// NewClient creates a new Scryfall API client with default configuration.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimiter: time.NewTicker(RateLimit),
	}
}

func (c *Client) request(endpoint string) (*http.Response, error) {
	start := time.Now()
	logger := logging.WithComponent("scryfall")

	// Rate limiting.
	<-c.rateLimiter.C

	req, err := http.NewRequest("GET", BaseURL+endpoint, nil)
	if err != nil {
		metrics.RecordAPIRequest(false, time.Since(start).Milliseconds())
		return nil, errors.NewNetworkError("failed to create HTTP request", err)
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json")

	logger.Debug("Making API request", "endpoint", endpoint)

	resp, err := c.httpClient.Do(req)
	responseTime := time.Since(start).Milliseconds()

	if err != nil {
		metrics.RecordAPIRequest(false, responseTime)
		logging.LogError(logger, errors.NewNetworkError("HTTP request failed", err), "API request failed")

		return nil, errors.NewNetworkError("failed to execute HTTP request", err)
	}

	logging.LogAPIRequest(endpoint, responseTime)

	if resp.StatusCode >= 400 {
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				logger.Warn("Failed to close response body", "error", closeErr)
			}
		}()

		var errResp Error
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			metrics.RecordAPIRequest(false, responseTime)
			return nil, errors.FromHTTPStatus(resp.StatusCode, fmt.Sprintf("HTTP error %d", resp.StatusCode))
		}

		metrics.RecordAPIRequest(false, responseTime)
		// Create MTGError for proper metrics tracking.
		mtgErr := errors.FromHTTPStatus(errResp.Status, errResp.Details)
		metrics.RecordError(mtgErr)

		return nil, errResp
	}

	metrics.RecordAPIRequest(true, responseTime)

	return resp, nil
}

// GetCardByName searches for a card by name using fuzzy matching.
func (c *Client) GetCardByName(name string) (*Card, error) {
	logger := logging.WithComponent("scryfall").With("card_name", name)

	if name == "" {
		return nil, errors.NewValidationError("card name cannot be empty")
	}

	endpoint := fmt.Sprintf("/cards/named?fuzzy=%s", url.QueryEscape(name))

	resp, err := c.request(endpoint)
	if err != nil {
		logging.LogError(logger, err, "Failed to request card by name")
		return nil, errors.NewAPIError("failed to fetch card by name", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Warn("Failed to close response body", "error", closeErr)
		}
	}()

	var card Card
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, errors.NewAPIError("failed to decode card response", err)
	}

	logger.Debug("Successfully retrieved card", "card_name", card.Name)

	return &card, nil
}

// GetCardByExactName searches for a card by exact name match.
func (c *Client) GetCardByExactName(name string) (*Card, error) {
	logger := logging.WithComponent("scryfall").With("card_name", name)

	if name == "" {
		return nil, errors.NewValidationError("card name cannot be empty")
	}

	endpoint := fmt.Sprintf("/cards/named?exact=%s", url.QueryEscape(name))

	resp, err := c.request(endpoint)
	if err != nil {
		logging.LogError(logger, err, "Failed to request card by exact name")
		return nil, errors.NewAPIError("failed to fetch card by exact name", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Warn("Failed to close response body", "error", closeErr)
		}
	}()

	var card Card
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, errors.NewAPIError("failed to decode card response", err)
	}

	logger.Debug("Successfully retrieved card by exact name", "card_name", card.Name)

	return &card, nil
}

// GetRandomCard returns a random Magic card.
func (c *Client) GetRandomCard() (*Card, error) {
	logger := logging.WithComponent("scryfall")
	endpoint := "/cards/random"

	resp, err := c.request(endpoint)
	if err != nil {
		logging.LogError(logger, err, "Failed to request random card")
		return nil, errors.NewAPIError("failed to fetch random card", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Warn("Failed to close response body", "error", closeErr)
		}
	}()

	var card Card
	if err := json.NewDecoder(resp.Body).Decode(&card); err != nil {
		return nil, errors.NewAPIError("failed to decode random card response", err)
	}

	logger.Debug("Successfully retrieved random card", "card_name", card.Name)

	return &card, nil
}

// SearchCards performs a full-text search for cards.
func (c *Client) SearchCards(query string) (*SearchResult, error) {
	logger := logging.WithComponent("scryfall")

	if query == "" {
		return nil, errors.NewValidationError("search query cannot be empty")
	}

	endpoint := fmt.Sprintf("/cards/search?q=%s", url.QueryEscape(query))

	resp, err := c.request(endpoint)
	if err != nil {
		logging.LogError(logger, err, "Failed to search cards")
		return nil, errors.NewAPIError("failed to search cards", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Warn("Failed to close response body", "error", closeErr)
		}
	}()

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.NewAPIError("failed to decode search response", err)
	}

	logger.Debug("Successfully searched cards", "query", query, "results", result.TotalCards)

	return &result, nil
}

// SearchCardFirst performs a search and returns the first result, useful for filtered card lookups.
func (c *Client) SearchCardFirst(query string) (*Card, error) {
	logger := logging.WithComponent("scryfall").With("query", query)

	if query == "" {
		return nil, errors.NewValidationError("search query cannot be empty")
	}

	// Add order by relevance and limit to 1 result for efficiency
	searchQuery := fmt.Sprintf("(%s) order:relevance", query)
	endpoint := fmt.Sprintf("/cards/search?q=%s", url.QueryEscape(searchQuery))

	resp, err := c.request(endpoint)
	if err != nil {
		logging.LogError(logger, err, "Failed to search for first card")
		return nil, errors.NewAPIError("failed to search for card", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Warn("Failed to close response body", "error", closeErr)
		}
	}()

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, errors.NewAPIError("failed to decode search response", err)
	}

	if result.TotalCards == 0 || len(result.Data) == 0 {
		return nil, errors.NewAPIError("no cards found matching query", fmt.Errorf("search returned no results"))
	}

	logger.Debug("Successfully found card via search", "card_name", result.Data[0].Name, "total_results", result.TotalCards)

	return &result.Data[0], nil
}

// Close stops the rate limiter ticker.
func (c *Client) Close() {
	if c.rateLimiter != nil {
		c.rateLimiter.Stop()
	}
}

// GetBestImageURL returns the highest quality image URL available for a card.
func (c *Card) GetBestImageURL() string {
	var imageUris map[string]string

	// For double-faced cards, prefer the first face.
	switch {
	case len(c.CardFaces) > 0 && c.CardFaces[0].ImageUris != nil:
		imageUris = c.CardFaces[0].ImageUris
	case c.ImageUris != nil:
		imageUris = c.ImageUris
	default:
		return ""
	}

	// Prefer highest quality images in order.
	imagePreference := []string{"png", "large", "normal", "small"}

	for _, format := range imagePreference {
		if url, exists := imageUris[format]; exists {
			return url
		}
	}

	// Return any available image if none of the preferred formats exist.
	for _, url := range imageUris {
		return url
	}

	return ""
}

// GetDisplayName returns the appropriate display name for the card.
func (c *Card) GetDisplayName() string {
	if c.Name != "" {
		return c.Name
	}

	// For multi-faced cards without a combined name.
	if len(c.CardFaces) > 0 {
		names := make([]string, len(c.CardFaces))
		for i, face := range c.CardFaces {
			names[i] = face.Name
		}

		return strings.Join(names, " // ")
	}

	return "Unknown Card"
}

// IsValidCard checks if the card has valid data for display.
func (c *Card) IsValidCard() bool {
	return c.Object == "card" && (c.Name != "" || len(c.CardFaces) > 0)
}

// HasImage checks if the card has at least one image available.
func (c *Card) HasImage() bool {
	return c.GetBestImageURL() != ""
}
