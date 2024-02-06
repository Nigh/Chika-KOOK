package main

import "encoding/json"

type KookCard struct {
	Card kCard
}

type kTheme string

const (
	kkPrimary   kTheme = "primary"
	kkSuccess   kTheme = "success"
	kkDanger    kTheme = "danger"
	kkWarning   kTheme = "warning"
	kkInfo      kTheme = "info"
	kkSecondary kTheme = "secondary"
	kkNone      kTheme = "none"
)

type kType0 string // card
type kType1 string // modules
type kType2 string // fields
type kSize string

const (
	kkLarge  kSize = "lg"
	kkMedium kSize = "md"
	kkSmall  kSize = "sm"
	kkXSmall kSize = "xs"
)
const (
	kkCard kType0 = "card"
)
const (
	kkHeader    kType1 = "header"
	kkSection   kType1 = "section"
	kkContext   kType1 = "context"
	kkDivider   kType1 = "divider"
	kkCountdown kType1 = "countdown"
	kkContainer kType1 = "container"
	kkFile      kType1 = "file"
)
const (
	kkPlaintext kType2 = "plain-text"
	kkImage     kType2 = "image"
	kkMarkdown  kType2 = "kmarkdown"
	kkParagraph kType2 = "paragraph"
)

type kkField struct {
	Type    kType2    `json:"type"`
	Content string    `json:"content,omitempty"`
	Src     string    `json:"src,omitempty"`
	Cols    int       `json:"cols,omitempty"`
	Fields  []kkField `json:"fields,omitempty"`
}

type kkModule struct {
	Type kType1 `json:"type,omitempty"`

	// header, section
	Text kkField `json:"text,omitempty"`

	// context, container
	Elements []kkField `json:"elements,omitempty"`

	// countdown
	Mode      string `json:"mode,omitempty"`
	StartTime int64  `json:"startTime,omitempty"`
	EndTime   int64  `json:"endTime,omitempty"`

	// file
	Title string `json:"title,omitempty"`
	Src   int    `json:"src,omitempty"`
	Size  int    `json:"size,omitempty"`
}

type kCard struct {
	Type    kType0     `json:"type"`
	Theme   kTheme     `json:"theme"`
	Size    kSize      `json:"size"`
	Modules []kkModule `json:"modules"`
}

func (card *KookCard) Init() *KookCard {
	card.Card.Type = kkCard
	card.Card.Size = kkLarge
	card.Card.Theme = kkPrimary
	return card
}
func (card *KookCard) AddModule(module kkModule) {
	card.Card.Modules = append(card.Card.Modules, module)
}

func (card *KookCard) AddModule_markdown(content string) {
	card.Card.Modules = append(card.Card.Modules, kkModule{
		Type: "section",
		Text: kkField{
			Type:    "kmarkdown",
			Content: content,
		},
	})
}
func (card *KookCard) AddModule_header(content string) {
	card.Card.Modules = append(card.Card.Modules, kkModule{
		Type: "header",
		Text: kkField{
			Type:    "plain-text",
			Content: content,
		},
	})
}
func (card *KookCard) AddModule_divider() {
	card.Card.Modules = append(card.Card.Modules, kkModule{
		Type: "divider",
	})
}
func (card *KookCard) String() string {
	jsons, _ := json.Marshal([]kCard{card.Card})
	return string(jsons)
}
