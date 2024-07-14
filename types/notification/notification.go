package notification

import "fmt"

type Image struct {
	URL    string  `json:"url"`
}

func (i Image) Display() string {
	return fmt.Sprintf("  URL: %s", i.URL)
}

type Notification struct {
	Type     string  `json:"type"` // "notification"
	Title    *string `json:"title"`
	Subtitle *string `json:"subtitle"`
	Content  *string `json:"content"`
	Image    *Image  `json:"image"`
}

func (n Notification) Display() string {
	return fmt.Sprintf("Title: %s\nSubtitle: %s\nContent: %s\nImage:\n%s", *n.Title, *n.Subtitle, *n.Content, n.Image.Display())
}
