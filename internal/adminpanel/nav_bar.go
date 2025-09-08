package adminpanel

import "fmt"

// NavBarItem represents an item in the navigation bar.
type NavBarItem struct {
	Name              string
	Link              string
	Bold              bool
	NavBarAppendSlash bool
}

// HTML returns the HTML representation of the navigation bar item.
func (i *NavBarItem) HTML() string {
    if i.Link != "" {
        return fmt.Sprintf(`<a class="nav-link" href="%s">%s</a>`, i.Link, i.Name)
    }
    if i.Bold {
        return fmt.Sprintf(`<span class="navbar-text fw-semibold me-2">%s</span>`, i.Name)
    }
    return fmt.Sprintf(`<span class="navbar-text me-2">%s</span>`, i.Name)
}

// NavBarGenerator defines a function type for generating navigation bar items.
type NavBarGenerator = func(ctx interface{}) NavBarItem
