package common

//BotConfig config for the bot
type BotConfig struct {
	BotKey         string `json:"bot_key"`
	UnsplashKey    string `json:"unsplash_key"`
	UnsplashSecret string `json:"unsplash_secret"`
}

//User represent an unsplash user
type User struct {
	ID                string `json:"id"`
	Username          string `json:"username"`
	Name              string `json:"name"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	InstagramUsername string `json:"instagram_username"`
	TwitterUsername   string `json:"twitter_username"`
	PortfolioURL      string `json:"portfolio_url"`
	ProfileImage      struct {
		Small  string `json:"small"`
		Medium string `json:"medium"`
		Large  string `json:"large"`
	} `json:"profile_image"`
}

//UnsplashPhoto represent an unsplash photo
type UnsplashPhoto struct {
	ID          string `json:"id"`
	CreatedAt   string `json:"created_at"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Color       string `json:"color"`
	Likes       int    `json:"likes"`
	LikedByUser bool   `json:"liked_by_user"`
	Description string `json:"description"`
	User        User   `json:"user"`
	URLs        struct {
		Raw     string `json:"raw"`
		Full    string `json:"full"`
		Regular string `json:"regular"`
		Small   string `json:"small"`
		Thumb   string `json:"thumb"`
	} `json:"urls"`
	Links struct {
		Self     string `json:"self"`
		HTML     string `json:"html"`
		Download string `json:"download"`
	} `json:"links"`
}

//SearchResult result for search from Unsplash
type SearchResult struct {
	Total      int             `json:"total"`
	TotalPages int             `json:"total_pages"`
	Results    []UnsplashPhoto `json:"results"`
}
