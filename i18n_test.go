package i18n_test

import (
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/httptest"
	i18n "github.com/gobuffalo/mw-i18n/v2"
	"github.com/stretchr/testify/require"
)

type User struct {
	FirstName string
	LastName  string
}

func app() *buffalo.App {
	app := buffalo.New(buffalo.Options{})

	r := render.New(render.Options{
		TemplatesFS: os.DirFS("templates"),
	})

	// Setup and use translations:
	t, err := i18n.New(os.DirFS("locales"), "en-US")
	if err != nil {
		log.Fatal(err)
	}
	// Setup URL prefix Language extractor
	t.LanguageExtractors = append(t.LanguageExtractors, i18n.URLPrefixLanguageExtractor)

	app.Use(t.Middleware())
	app.GET("/", func(c buffalo.Context) error {
		return c.Render(200, r.HTML("index.html"))
	})
	app.GET("/plural", func(c buffalo.Context) error {
		return c.Render(200, r.HTML("plural.html"))
	})
	app.GET("/format", func(c buffalo.Context) error {
		usersList := make([]User, 0)
		usersList = append(usersList, User{"Mark", "Bates"})
		usersList = append(usersList, User{"Chuck", "Berry"})
		c.Set("Users", usersList)
		return c.Render(200, r.HTML("format.html"))
	})
	app.GET("/collision", func(c buffalo.Context) error {
		return c.Render(200, r.HTML("collision.html"))
	})
	app.GET("/localized", func(c buffalo.Context) error {
		return c.Render(200, r.HTML("localized_view.html"))
	})
	app.GET("/languages-list", func(c buffalo.Context) error {
		return c.Render(200, r.JSON(t.AvailableLanguages()))
	})
	app.GET("/refresh", func(c buffalo.Context) error {
		// This flash will be displayed in english
		c.Flash().Add("success", t.Translate(c, "refresh-success"))

		// Change lang to fr-fr
		c.Cookies().Set("lang", "fr-fr", time.Minute)
		t.Refresh(c, "fr-fr")

		// This flash will be displayed in french
		c.Flash().Add("success", t.Translate(c, "refresh-success"))
		return c.Render(200, r.HTML("refresh.html"))
	})
	// Disable i18n middleware
	noI18n := func(c buffalo.Context) error {
		return c.Render(200, r.HTML("localized_view.html"))
	}
	app.Middleware.Skip(t.Middleware(), noI18n)
	app.GET("/localized-disabled", noI18n)
	app.GET("/{lang:fr|en}/index", func(c buffalo.Context) error {
		return c.Render(200, r.HTML("index.html"))
	})
	return app
}

func Test_i18n(t *testing.T) {
	r := require.New(t)

	w := httptest.New(app())
	res := w.HTML("/").Get()
	r.Equal("Hello, World!", strings.TrimSpace(res.Body.String()))
}

func Test_i18n_fr(t *testing.T) {
	r := require.New(t)

	w := httptest.New(app())
	req := w.HTML("/")
	// Set language as "french"
	req.Headers["Accept-Language"] = "fr-fr"
	res := req.Get()
	r.Equal("Bonjour à tous !", strings.TrimSpace(res.Body.String()))
}

func Test_i18n_plural(t *testing.T) {
	r := require.New(t)

	w := httptest.New(app())
	res := w.HTML("/plural").Get()
	r.True(eq("Hello, alone!\nHello, 5 people!", strings.TrimSpace(res.Body.String())))
}

func Test_i18n_plural_fr(t *testing.T) {
	r := require.New(t)

	w := httptest.New(app())
	req := w.HTML("/plural")
	// Set language as "french"
	req.Headers["Accept-Language"] = "fr-fr"
	res := req.Get()
	r.True(eq("Bonjour, tout seul !\nBonjour, 5 personnes !", strings.TrimSpace(res.Body.String())))
}

func Test_i18n_format(t *testing.T) {
	r := require.New(t)

	w := httptest.New(app())
	res := w.HTML("/format").Get()
	r.True(eq("Hello Mark!\n\n\t* Mr. Mark Bates\n\n\t* Mr. Chuck Berry\n", res.Body.String()))
}

func Test_i18n_format_fr(t *testing.T) {
	r := require.New(t)

	w := httptest.New(app())
	req := w.HTML("/format")
	// Set language as "french"
	req.Headers["Accept-Language"] = "fr-fr"
	res := req.Get()
	r.True(eq("Bonjour Mark !\n\n\t* M. Mark Bates\n\n\t* M. Chuck Berry\n", res.Body.String()))
}

func Test_i18n_Localized_View(t *testing.T) {
	r := require.New(t)

	w := httptest.New(app())
	// Test with complex Accept-Language
	req := w.HTML("/localized")
	req.Headers["Accept-Language"] = "en-UK,en-US;q=0.5"
	res := req.Get()
	r.Equal("Hello!", strings.TrimSpace(res.Body.String()))

	// Test priority
	req.Headers["Accept-Language"] = "fr,en-US"
	res = req.Get()
	r.Equal("Bonjour !", strings.TrimSpace(res.Body.String()))

	// Test fallback
	req.Headers["Accept-Language"] = "ru"
	res = req.Get()
	r.Equal("Default", strings.TrimSpace(res.Body.String()))

	// Test i18n disabled
	req = w.HTML("/localized-disabled")
	req.Headers["Accept-Language"] = "en-UK,en-US;q=0.5"
	res = req.Get()
	r.Equal("Default", strings.TrimSpace(res.Body.String()))
}

func Test_i18n_collision(t *testing.T) {
	r := require.New(t)

	w := httptest.New(app())
	res := w.HTML("/collision").Get()
	r.Equal("Collision OK", strings.TrimSpace(res.Body.String()))
}

func Test_i18n_availableLanguages(t *testing.T) {
	r := require.New(t)

	w := httptest.New(app())
	res := w.HTML("/languages-list").Get()
	r.Equal("[\"en-us\",\"fr-fr\"]", strings.TrimSpace(res.Body.String()))
}

func Test_i18n_URL_prefix(t *testing.T) {
	r := require.New(t)

	w := httptest.New(app())
	req := w.HTML("/fr/index")
	res := req.Get()
	r.Equal("Bonjour à tous !", strings.TrimSpace(res.Body.String()))

	req = w.HTML("/en/index")
	res = req.Get()
	r.Equal("Hello, World!", strings.TrimSpace(res.Body.String()))
}

func Test_i18n_TranslateWithLang(t *testing.T) {
	r := require.New(t)

	_ = httptest.New(app())
	transl := i18n.Translator{}

	// Test English
	lang := "en"
	original := "greeting"
	want := "Hello, World!"
	res, err := transl.TranslateWithLang(lang, original)
	r.NoError(err)
	r.Equal(want, res)

	// Test French
	lang = "fr-fr"
	want = "Bonjour à tous !"
	res, err = transl.TranslateWithLang(lang, original)
	r.NoError(err)
	r.Equal(want, res)

	// Test French singular
	original = "greeting-plural"
	want = "Bonjour, tout seul !"
	count := 1
	res, err = transl.TranslateWithLang(lang, original, count)
	r.NoError(err)
	r.Equal(want, res)

	// Test French plural
	want = "Bonjour, 5 personnes !"
	count = 5
	res, err = transl.TranslateWithLang(lang, original, count)
	r.NoError(err)
	r.Equal(want, res)

	// Test French format-loop
	original = "test-format-loop"
	want = "M. Mark Bates"
	params := struct{ FirstName, LastName string }{"Mark", "Bates"}
	res, err = transl.TranslateWithLang(lang, original, params)
	r.NoError(err)
	r.Equal(want, res)
}

func Test_Refresh(t *testing.T) {
	r := require.New(t)

	w := httptest.New(app())
	req := w.HTML("/refresh")
	res := req.Get()
	r.Equal("success: Language changed!#success: Langue modifiée !#", strings.TrimSpace(res.Body.String()))
}

func eq(a, b string) bool {
	clean := func(s string) string {
		s = strings.TrimSpace(strings.Replace(s, "\n", "", -1))
		s = strings.Replace(s, "\r", "", -1)
		return s
	}
	return clean(a) == clean(b)
}
