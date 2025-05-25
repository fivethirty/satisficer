package section

import (
	"slices"
	"sort"
	"time"
)

type Page struct {
	URL       string
	Title     string
	CreatedAt time.Time
	UpdatedAt *time.Time
	Content   string
}

type Section struct {
	Pages []Page
}

func (s Section) ByTitle() Section {
	sort.SliceStable(s.Pages, func(i, j int) bool {
		return s.Pages[i].Title < s.Pages[j].Title
	})
	return s
}

func (s Section) ByCreatedAt() Section {
	sort.SliceStable(s.Pages, func(i, j int) bool {
		return s.Pages[i].CreatedAt.Before(s.Pages[j].CreatedAt)
	})
	return s
}

func (s Section) Reverse() Section {
	slices.Reverse(s.Pages)
	return s
}
