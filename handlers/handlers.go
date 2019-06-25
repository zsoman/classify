package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"../sqliteDB"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type User struct {
	Email    string
	Password string
	FName    string
	LName    string
}

type Bookmark struct {
	Title       string
	Link        string
	Tags        string
	Categories  string
	Description string
	CDate       string
	MDate       string
	User        string
}

type Edit_bookmark struct {
	Email  		string
	BookmarkID  int
	Title       string
	Link  		string
	Tags 		string
	Description	string
}

var bookmarksPerPage = 5
var tagCloudMaxSize = 40
var InUser User
var NewUser User
var NewBookmark database.Bookmark
var Bookmarks []database.Bookmark
var LoggedIn = false
var LoginError = false
var RegisterError = false


func HomeHandlerEmpty(w http.ResponseWriter, r *http.Request) {
	log.Info("Home handler is started from handlers package.")
	if !LoggedIn {
		if database.IsEmptyUsers() {
			log.Info("Home handler redirected to login html.")
			http.Redirect(w, r, "/login", http.StatusFound)
		} else {
			log.Info("Home handler redirected to register html.")
			http.Redirect(w, r, "/register", http.StatusFound)
		}
	} else {
		log.Info("Home handler redirected to bookmarks html.")
		http.Redirect(w, r, "/bookmarks/"+strings.Split(InUser.Email, "@")[0], http.StatusFound)
	}

}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Login handler function started from handlers package.")
	if LoggedIn {
		LoggedIn = false
		log.Error(fmt.Sprintf("%s user is successfully logged out.",InUser.Email))
		InUser = User{}
		LoginError = false
	}
	if !LoginError {
		if r.Method == "GET" {
			t, _ := template.ParseFiles("Required/HTML templates/home_login.html")
			t.Execute(w, nil)
		} else {
			r.ParseForm()
			InUser = User{Email: r.FormValue("email"), Password: r.FormValue("pwd")}
			if database.Login(InUser.Email, InUser.Password) {
				log.Info(fmt.Sprintf("The %s user successfully loegged in.",r.FormValue("email")))
				LoggedIn = true
				Bookmarks = database.AllBookmarks(InUser.Email)
				http.Redirect(w, r, "/bookmarks/"+strings.Split(InUser.Email, "@")[0], http.StatusFound)
			} else {
				LoginError = true
				log.Info(fmt.Sprintf("The %s user couldn`t loegg in.",r.FormValue("email")))
				http.Redirect(w, r, "/login", http.StatusFound)
			}
		}
	} else {
		log.Info("Login error handler is started from handlers package.")
		LoginError = true
		if r.Method == "GET" {
			t, _ := template.ParseFiles("Required/HTML templates/home_login_error.html")
			t.Execute(w, nil)
		} else {
			r.ParseForm()
			InUser = User{Email: r.FormValue("email"), Password: r.FormValue("pwd")}
			if database.Login(InUser.Email, InUser.Password) {
				log.Info(fmt.Sprintf("The %s user successfully loegged in.",r.FormValue("email")))
				LoggedIn = true
				http.Redirect(w, r, "/bookmarks/"+strings.Split(InUser.Email, "@")[0], http.StatusFound)
			} else {
				log.Info(fmt.Sprintf("The %s user couldn`t loegg in.",r.FormValue("email")))
				http.Redirect(w, r, "/login", http.StatusFound)
			}
		}
	}
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Register handler is started from handlers package.")
	if !RegisterError {
		t := template.New("register")
		if r.Method == "GET" {
			t, _ = template.ParseFiles("Required/HTML templates/home_register.html")
			t.Execute(w, nil)
		} else {
			r.ParseForm()
			NewUser = User{Email: r.FormValue("email"), Password: r.FormValue("pwd"),
				FName: r.FormValue("fname"), LName: r.FormValue("lname")}
			if database.Register(NewUser.Email, NewUser.Password, NewUser.FName, NewUser.LName) {
				InUser = NewUser
				log.Info(fmt.Sprintf("%s new user is successfully created",r.FormValue("email")))
				LoggedIn = true
				http.Redirect(w, r, "/bookmarks/"+strings.Split(InUser.Email, "@")[0], http.StatusFound)
			}
			RegisterError = true
			log.Info(fmt.Sprintf("%s user already exists.",r.FormValue("email")))
			http.Redirect(w, r, "/register", http.StatusFound)
		}
	} else {
		t := template.New("register")
		if r.Method == "GET" {
			t, _ = template.ParseFiles("Required/HTML templates/home_register_error.html")
			t.Execute(w, nil)
		} else {
			r.ParseForm()
			NewUser = User{Email: r.FormValue("email"), Password: r.FormValue("pwd"),
				FName: r.FormValue("fname"), LName: r.FormValue("lname")}
			if database.Register(NewUser.Email, NewUser.Password, NewUser.FName, NewUser.LName) {
				InUser = NewUser
				log.Info(fmt.Sprintf("%s new user is successfully created",r.FormValue("email")))
				LoggedIn = true
				http.Redirect(w, r, "/bookmarks/"+strings.Split(InUser.Email, "@")[0], http.StatusFound)
			}
			log.Info(fmt.Sprintf("%s user already exists.",r.FormValue("email")))
			http.Redirect(w, r, "/register", http.StatusFound)
		}
	}
}

func LoggedHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Logged handler is started from handlers package.")
	if r.Method == "GET" {
		t := template.New("logged")
		t, _ = template.ParseFiles("Required/HTML templates/user_interface.html")
		t.Execute(w, InUser)
		t2 := template.New("savHTML")
		t2_1 := template.New("savHTML1")
		t2_2 := template.New("savHTML2")
		t3 := template.New("tadCloud")
		t4 := template.New("end")
		savHTML := ""

		vars := mux.Vars(r)
		oldalSzam, _ := strconv.Atoi(vars["page"])
		tagCloudNr, _ := strconv.Atoi(vars["tagcloud"])
		allBookmarksList := database.AllBookmarks(InUser.Email)
		var last int
		if len(allBookmarksList) > bookmarksPerPage {
			if oldalSzam != 0 {
				savHTML = savHTML + fmt.Sprintf("\t<a class=\"hvr-bounce-to-top\" href=\"/bookmarks/%s/page=%d\"><strong>Previous</strong></a> | \n\t", strings.Split(InUser.Email, "@")[0], oldalSzam-1)
			}

			for i := 0; i < len(allBookmarksList)/bookmarksPerPage; i++ {
				if i != 0 {
					savHTML = savHTML + " | \n\t"
				}
				if i==oldalSzam{
					savHTML = savHTML + fmt.Sprintf("\t<a class=\"hvr-bounce-to-top2\"><strong>%d-%d</strong></a>", (i*bookmarksPerPage)+1, ((i+1)*bookmarksPerPage))
				} else {
					savHTML = savHTML + fmt.Sprintf("\t<a class=\"hvr-bounce-to-top\" href=\"/bookmarks/%s/page=%d\"><strong>%d-%d</strong></a>", strings.Split(InUser.Email, "@")[0], i, (i*bookmarksPerPage)+1, ((i+1)*bookmarksPerPage))
				}
				last = i
			}
			if len(allBookmarksList) > bookmarksPerPage && len(allBookmarksList)%bookmarksPerPage != 0 {
				if oldalSzam==(len(allBookmarksList)/bookmarksPerPage) {
					savHTML = savHTML + fmt.Sprintf(" | \n\t\t<a class=\"hvr-bounce-to-top2\"><strong>%d-%d</strong></a>", ((last+1)*bookmarksPerPage)+1, ((last+1)*bookmarksPerPage)+len(allBookmarksList)%bookmarksPerPage)
 				} else {
					savHTML = savHTML + fmt.Sprintf(" | \n\t\t<a class=\"hvr-bounce-to-top\" href=\"/bookmarks/%s/page=%d\"><strong>%d-%d</strong></a>", strings.Split(InUser.Email, "@")[0], last+1, ((last+1)*bookmarksPerPage)+1, ((last+1)*bookmarksPerPage)+len(allBookmarksList)%bookmarksPerPage)
				}
			}
			if oldalSzam < len(allBookmarksList)/bookmarksPerPage {
				savHTML = savHTML + fmt.Sprintf(" | \n\t\t<a class=\"hvr-bounce-to-top\" href=\"/bookmarks/%s/page=%d\"><strong>Next</strong></a>", strings.Split(InUser.Email, "@")[0], oldalSzam+1)
			}

		}

		tagCloud := fmt.Sprintf(`	
	</div>
</div>
		
<div class="right-column">
	<div class="tagcloud_header_box">
		<a class="tagcloud_header" href="/bookmarks/%s/page=%d/tagcloud=0">Tag Cloud</a>
		<a class="tagcloud_header_numbers" href="/bookmarks/%s/page=%d/tagcloud=20">20</a>
		<a class="tagcloud_header_numbers" href="/bookmarks/%s/page=%d/tagcloud=10">10</a>
		<a class="tagcloud_header_numbers" href="/bookmarks/%s/page=%d/tagcloud=5">5</a>
		<a class="tagcloud_header_numbers" href="/bookmarks/%s/page=%d/tagcloud=2">2</a>
	</div>
	<br>`, strings.Split(InUser.Email, "@")[0], oldalSzam, strings.Split(InUser.Email, "@")[0], oldalSzam, strings.Split(InUser.Email, "@")[0], oldalSzam, strings.Split(InUser.Email, "@")[0], oldalSzam, strings.Split(InUser.Email, "@")[0], oldalSzam)
		tagCloudMap := database.TagsCount(InUser.Email, tagCloudNr)
		var tagCloudKeys []string
		for k := range tagCloudMap {
			tagCloudKeys = append(tagCloudKeys, k)
		}
		sort.Strings(tagCloudKeys)
		for _, k := range tagCloudKeys {
			var size int
			if tagCloudMaxSize < 15+3*tagCloudMap[k] {
				size = tagCloudMaxSize
			} else {
				size = 15 + 3*tagCloudMap[k]
			}
			tagCloud += fmt.Sprintf(`
	<a class= "hvr-bob" href="/bookmarks/%s/page=0/tag=%s/tagcloud=0" style="font-size:%dpx;">%s </a>`,
				strings.Split(InUser.Email, "@")[0], k, size, k)
		}

		savHTML = savHTML + "\n\t</div>\n"
		t2_1.Parse(`<div id="alatta-wrap">
	<div class="left-column">
	<div id="felso_lepteto">
	`)
		t2_1.Execute(w, nil)
		t2.Parse(savHTML)
		t2.Execute(w, nil)
			
		var online bool
		if isOnline() {
			online = true
		} else {
			online = false
		}
		for nr, i := range allBookmarksList {
			if nr >= (bookmarksPerPage*oldalSzam) && nr < (bookmarksPerPage*(oldalSzam+1)) {
				p := template.New("bookmark")
				p2 := template.New("tags")
				tmp := i
				time1, _ := time.Parse("2006-01-02 15:04", tmp.CDate)
				time2, _ := time.Parse("2006-01-02 15:04", tmp.MDate)
				tmp.CDate = timeAgo(time1)
				tmp.MDate = timeAgo(time2)
				var display_bookmark string
				if online {
					display_bookmark = fmt.Sprintf("\n\n\t<div class=\"flip-container\">\n\t\t<div class=\"flipper\" onclick=\"this.classList.toggle('flipped')\">\n\t\t\t<div class=\"front\">\n\t\t\t\t<div class=\"front-side\">\n\t\t\t\t\t<a href=\"%s\">\n\t\t\t\t\t\t<img class=\"honlap_kep_front\" src=\"http://localhost:8080/Required/Images/Screenshots/screenshot_%d.jpg\" alt=\"Website screenshot\">\n\t\t\t\t\t</a>\n\t\t\t\t\t<div class=\"leiras\">\n\t\t\t\t\t\t<a class=\"flip_cim\" href=\"%s\">%s</a>\n\t\t\t\t\t\t<p class=\"flip_description\">%s</p>", tmp.Link, tmp.ID, tmp.Link, tmp.Title, tmp.Description)
				} else {
					display_bookmark = fmt.Sprintf("\n\n\t<div class=\"flip-container\">\n\t\t<div class=\"flipper\" onclick=\"this.classList.toggle('flipped')\">\n\t\t\t<div class=\"front\">\n\t\t\t\t<div class=\"front-side\">\n\t\t\t\t\t<a href=\"http://localhost:8080/Required/Websites/%d/%d.html\">\n\t\t\t\t\t\t<img class=\"honlap_kep_front\" src=\"http://localhost:8080/Required/Images/Screenshots/screenshot_%d.jpg\" alt=\"Website screenshot\">\n\t\t\t\t\t</a>\n\t\t\t\t\t<div class=\"leiras\">\n\t\t\t\t\t\t<a class=\"flip_cim\" href=\"http://localhost:8080/Required/Websites/%d/%d.html\">%s - offline</a>\n\t\t\t\t\t\t<p class=\"flip_description\">%s</p>", tmp.ID, tmp.ID, tmp.ID, tmp.ID, tmp.ID, tmp.Title, tmp.Description)					
				}
				bookmarkTags := "\n\t\t\t\t\t\t<div class=\"flip_all_tags_front\">"
				bookmarkTags2 := "\n\t\t\t\t\t\t<div class=\"flip_all_tags_back\">"
				for _, tags := range strings.Split(tmp.Tags, ";") {
					bookmarkTags += fmt.Sprintf("\n\t\t\t\t\t\t\t<a class=\"flip_tag\" href=\"/bookmarks/%s/page=0/tag=%s\"><strong>%s</strong></a>", strings.Split(InUser.Email, "@")[0], tags, tags)
					bookmarkTags2 += fmt.Sprintf("\n\t\t\t\t\t\t\t<a class=\"flip_tag\" href=\"/bookmarks/%s/page=0/tag=%s\"><strong>%s</strong></a>", strings.Split(InUser.Email, "@")[0], tags, tags)
				}

				bookmarkTags += fmt.Sprintf("\n\t\t\t\t\t\t</div>\n\t\t\t\t\t</div>\n\t\t\t\t</div>\n\t\t\t</div>")
				if online {
					bookmarkTags += fmt.Sprintf("\n\t\t\t<div class=\"back-side\">\n\t\t\t\t<a href=\"%s\">\n\t\t\t\t\t<img class=\"honlap_kep_front\" src=\"http://localhost:8080/Required/Images/Screenshots/screenshot_%d.jpg\" alt=\"Website screenshot\">\n\t\t\t\t</a>\n\t\t\t\t<div class=\"leiras\">\n\t\t\t\t\t<a class=\"flip_cim\" href=\"%s\">%s</a>", tmp.Link, tmp.ID, tmp.Link, tmp.Title)
				} else {
					bookmarkTags += fmt.Sprintf("\n\t\t\t<div class=\"back-side\">\n\t\t\t\t<a href=\"http://localhost:8080/Required/Websites/%d/%d.html\">\n\t\t\t\t\t<img class=\"honlap_kep_front\" src=\"http://localhost:8080/Required/Images/Screenshots/screenshot_%d.jpg\" alt=\"Website screenshot\">\n\t\t\t\t</a>\n\t\t\t\t<div class=\"leiras\">\n\t\t\t\t\t<a class=\"flip_cim\" href=\"http://localhost:8080/Required/Websites/%d/%d.html\">%s - offline</a>", tmp.ID, tmp.ID, tmp.ID, tmp.ID, tmp.ID, tmp.Title)					
				}
				bookmarkTags += bookmarkTags2 + "\n\t\t\t\t\t</div>"
				bookmarkTags += fmt.Sprintf("\n\t\t\t\t\t<p class=\"flip_creationdate\">Created on: %s<br>Last modified on: %s</p>", tmp.CDate, tmp.MDate)
				bookmarkTags += fmt.Sprintf("\n\t\t\t\t\t<a href='/edit_bookmark/bookmarkID=%d'><img class=\"flip_edit\" src=\"http://localhost:8080/Required/Images/Buttons/edit.png\" alt=\"Edit bookmark\"></a>", tmp.ID)
				bookmarkTags += fmt.Sprintf("\n\t\t\t\t\t<a href='/delete_bookmark/bookmarkID=%d'><img class=\"flip_delete\" src=\"http://localhost:8080/Required/Images/Buttons/delete.png\" alt=\"Delete bookmark\"></a>", tmp.ID)
				bookmarkTags += "\n\t\t\t\t</div>\n\t\t\t</div>\n\t\t</div>\n\t</div>"
				p.Parse(display_bookmark)
				p.Execute(w, nil)
				p2.Parse(bookmarkTags)
				p2.Execute(w, nil)
			}
		}
		savHTMLalso := "\n\t<div id=\"also_lepteto\">\n" + savHTML
		t2_2.Parse(savHTMLalso)
		t2_2.Execute(w, nil)
		tagCloud += "\n</div>"
		t3.Parse(tagCloud)
		t3.Execute(w, nil)
		t4.Parse("\n</div>\n</body>\n</html>")
		t4.Execute(w, nil)

		r.ParseForm()
	}
	if r.Method == "POST" {
		search := r.FormValue("search")
		http.Redirect(w, r, "/search/string="+search+"/page=0", http.StatusFound)
	}
}

func NewBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	t := template.New("register")
	log.Info("New bookmark handler is started.")
	if r.Method == "GET" {
		t, _ = template.ParseFiles("Required/HTML templates/new_bookmark.html")
		t.Execute(w, InUser)
		r.ParseForm()
	} else {
		r.ParseForm()
		log.Info("New bookmark form filled data.")
		t := time.Now().UTC().Unix()
		var title string
		var link string
		var tags string
		var categories string
		var description string
		if r.FormValue("title") == "" {
			tmp := strings.Split(r.FormValue("link"), "/")
			if len(tmp) == 1 {
				title = tmp[0]
			} else {
				if len(tmp[len(tmp)-1]) > 1 {
					title = tmp[len(tmp)-1]
				} else {
					title = tmp[len(tmp)-2]
				}
			}

		} else {
			title = r.FormValue("title")
		}
		if r.FormValue("tags") == "" {
			tags = " "
		} else {
			tags = r.FormValue("tags")
		}
		if r.FormValue("link") == "" {
			link = " "
		} else {
			link = r.FormValue("link")
		}
		if r.FormValue("categories") == "" {
			categories = " "
		} else {
			categories = r.FormValue("categories")
		}
		if r.FormValue("description") == "" {
			description = " "
		} else {
			description = r.FormValue("description")
		}
		if tags != " " {
			w := strings.FieldsFunc(tags, func(r rune) bool {
				switch r {
				case ',', '.' /*, ' '*/, ';':
					return true
				}
				return false
			})
			for nr, i := range w {
				w[nr] = strings.TrimSpace(i)
			}
			tags = strings.Join(w, ";")
		}

		NewBookmark = database.Bookmark{Title: title, Link: link,
			Tags: tags, Categories: categories, Description: description,
			CDate: strconv.FormatInt(t, 10) /*.Format("2006-01-02 15:04")*/, MDate: strconv.FormatInt(t, 10), /*.Format("2006-01-02 15:04")*/
			User: InUser.Email}
		database.NewBookmark(NewBookmark.Title, NewBookmark.Link, NewBookmark.Tags,
			NewBookmark.Categories, NewBookmark.Description, NewBookmark.CDate,
			NewBookmark.MDate, NewBookmark.User)
		log.Info(NewBookmark)
		Bookmarks = append(Bookmarks, NewBookmark)
		http.Redirect(w, r, "/bookmarks/"+strings.Split(InUser.Email, "@")[0], http.StatusFound)
	}
}

func EditBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	t := template.New("editBookmark")
	log.Info("Edit bookmark handler is started.")
	t, _ = template.ParseFiles("Required/HTML templates/edit_bookmark.html")
	vars := mux.Vars(r)
	bookmarkID, _ := strconv.Atoi(vars["bookmarkID"])
	bookmark := database.RetrieveOneBookmark(bookmarkID)
	if r.Method == "GET" {
		editBookmark:=Edit_bookmark{Email: strings.Split(InUser.Email, "@")[0], 
				BookmarkID: bookmarkID, Title: bookmark.Title, Link: bookmark.Link, 
				Tags: bookmark.Tags, Description: bookmark.Description}
		t.Execute(w, editBookmark)
	} else {
		r.ParseForm()
		log.Info("Edit bookmark form filled data.")
		t := time.Now().UTC().Unix()
		var title string
		var link string
		var tags string
		var categories string
		var description string

		if r.FormValue("title") == "" {
			tmp := strings.Split(r.FormValue("link"), "/")
			if len(tmp) == 1 {
				title = tmp[0]
			} else {
				if len(tmp[len(tmp)-1]) > 1 {
					title = tmp[len(tmp)-1]
				} else {
					title = tmp[len(tmp)-2]
				}
			}

		} else {
			title = r.FormValue("title")
		}
		if r.FormValue("tags") == "" {
			tags = " "
		} else {
			tags = r.FormValue("tags")
		}
		if r.FormValue("link") == "" {
			link = " "
		} else {
			link = r.FormValue("link")
		}
		if r.FormValue("categories") == "" {
			categories = " "
		} else {
			categories = r.FormValue("categories")
		}
		if r.FormValue("description") == "" {
			description = " "
		} else {
			description = r.FormValue("description")
		}
		if tags != " " {
			w := strings.FieldsFunc(tags, func(r rune) bool {
				switch r {
				case ',', '.' /*, ' '*/, ';':
					return true
				}
				return false
			})
			for nr, i := range w {
				w[nr] = strings.TrimSpace(i)
			}
			tags = strings.Join(w, ";")
		}

		NewBookmark = database.Bookmark{Title: title, Link: link,
			Tags: tags, Categories: categories, Description: description, CDate: bookmark.CDate,
			MDate: strconv.FormatInt(t, 10) /*.Format("2006-01-02 15:04")*/, User: InUser.Email}

		tmpCdate, _ := strconv.Atoi(NewBookmark.CDate)
		database.EditBookmark(bookmarkID, tmpCdate, NewBookmark.Title, NewBookmark.Link, NewBookmark.Tags,
			NewBookmark.Categories, NewBookmark.Description, NewBookmark.MDate, NewBookmark.User)
		http.Redirect(w, r, "/bookmarks/"+strings.Split(InUser.Email, "@")[0], http.StatusFound)
	}
}

func DeleteBookmarkHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Edit bookmark handler is started.")
	vars := mux.Vars(r)
	bookmarkID, _ := strconv.Atoi(vars["bookmarkID"])
	database.DeleteBookmark(bookmarkID)

	http.Redirect(w, r, "/bookmarks/"+strings.Split(InUser.Email, "@")[0], http.StatusFound)
}

func TagsHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("Tag handler is started, with %s tag.")
	t := template.New("tag")
	t, _ = template.ParseFiles("Required/HTML templates/user_interface.html")
	t.Execute(w, InUser)
	t2 := template.New("savHTML")
	t2_1 := template.New("savHTML1")
	t2_2 := template.New("savHTML2")
	t3 := template.New("tadCloud")
	t4 := template.New("end")
	savHTML := ""

	vars := mux.Vars(r)
	oldalSzam, _ := strconv.Atoi(vars["page"])
	tagCloudNr, _ := strconv.Atoi(vars["tagcloud"])
	tag := vars["tag"]
	allBookmarksList := database.TagsBookmarks(InUser.Email, tag)
	var last int
	if len(allBookmarksList) > bookmarksPerPage {
		if oldalSzam != 0 {
			savHTML = savHTML + fmt.Sprintf("\t\t<a class=\"hvr-bounce-to-top\" href=\"/bookmarks/%s/page=%d/tag=%s\"><strong>Previous</strong></a> | \n\t", strings.Split(InUser.Email, "@")[0], oldalSzam-1, tag)
		}

		for i := 0; i < len(allBookmarksList)/bookmarksPerPage; i++ {
			if i != 0 {
				savHTML = savHTML + " | \n\t"
			}
			if i==oldalSzam{
				savHTML = savHTML + fmt.Sprintf("\t\t<a class=\"hvr-bounce-to-top2\"><strong>%d-%d</strong></a>", (i*bookmarksPerPage)+1, ((i+1)*bookmarksPerPage))
			} else {
				savHTML = savHTML + fmt.Sprintf("\t\t<a class=\"hvr-bounce-to-top\" href=\"/bookmarks/%s/page=%d/tag=%s\"><strong>%d-%d</strong></a>", strings.Split(InUser.Email, "@")[0], i, tag, (i*bookmarksPerPage)+1, ((i+1)*bookmarksPerPage))
			}
			last = i
		}
		if len(allBookmarksList) > bookmarksPerPage && len(allBookmarksList)%bookmarksPerPage != 0 {
			if oldalSzam==(len(allBookmarksList)/bookmarksPerPage) {
				savHTML = savHTML + fmt.Sprintf(" | \n\t\t\t<a class=\"hvr-bounce-to-top2\"><strong>%d-%d</strong></a>", ((last+1)*bookmarksPerPage)+1, ((last+1)*bookmarksPerPage)+len(allBookmarksList)%bookmarksPerPage)
 			} else {
				savHTML = savHTML + fmt.Sprintf(" | \n\t\t\t<a class=\"hvr-bounce-to-top\" href=\"/bookmarks/%s/page=%d/tag=%s\"><strong>%d-%d</strong></a>", strings.Split(InUser.Email, "@")[0], last+1, tag, ((last+1)*bookmarksPerPage)+1, ((last+1)*bookmarksPerPage)+len(allBookmarksList)%bookmarksPerPage)
			}
		}
		if oldalSzam < len(allBookmarksList)/bookmarksPerPage {
			savHTML = savHTML + fmt.Sprintf(" | \n\t\t\t<a class=\"hvr-bounce-to-top\" href=\"/bookmarks/%s/page=%d/tag=%s\"><strong>Next</strong></a>", strings.Split(InUser.Email, "@")[0], oldalSzam+1, tag)
		}
	}

	tagCloud := fmt.Sprintf(`
	</div>
</div>
		
	<div class="right-column">
		<div class="tagcloud_header_box">
			<a class="tagcloud_header" href="/bookmarks/%s/page=%d/tag=%s/tagcloud=0">Tag Cloud</a>
			<a class="tagcloud_header_numbers" href="/bookmarks/%s/page=%d/tag=%s/tagcloud=20">20</a>
			<a class="tagcloud_header_numbers" href="/bookmarks/%s/page=%d/tag=%s/tagcloud=10">10</a>
			<a class="tagcloud_header_numbers" href="/bookmarks/%s/page=%d/tag=%s/tagcloud=5">5</a>
			<a class="tagcloud_header_numbers" href="/bookmarks/%s/page=%d/tag=%s/tagcloud=2">2</a>
		</div>
		<br>`, strings.Split(InUser.Email, "@")[0], oldalSzam, tag, strings.Split(InUser.Email, "@")[0], oldalSzam, tag, strings.Split(InUser.Email, "@")[0], oldalSzam, tag, strings.Split(InUser.Email, "@")[0], oldalSzam, tag, strings.Split(InUser.Email, "@")[0], oldalSzam, tag)
	tagCloudMap := database.TagsCount(InUser.Email, tagCloudNr)
	var tagCloudKeys []string
	for k := range tagCloudMap {
		tagCloudKeys = append(tagCloudKeys, k)
	}
	sort.Strings(tagCloudKeys)
	for _, k := range tagCloudKeys {
		var size int
		if tagCloudMaxSize < 15+3*tagCloudMap[k]{
			size = tagCloudMaxSize
		} else {
			size = 15 + 3*tagCloudMap[k]
		}
		tagCloud += fmt.Sprintf(`	<a class= "hvr-bob" href="/bookmarks/%s/page=0/tag=%s/tagcloud=0" style="font-size:%dpx;">%s </a>
		`,strings.Split(InUser.Email, "@")[0], k, size, k)
	}

	savHTML = savHTML + "\n\t\t</div>\n"
		t2_1.Parse(`<div id="alatta-wrap">
	<div class="left-column">
	<div id="felso_lepteto">
	`)
	t2_1.Execute(w, nil)
	t2.Parse(savHTML)
	t2.Execute(w, nil)

	var online bool
		if isOnline() {
			online = true
		} else {
			online = false
		}
	for nr, i := range allBookmarksList {
		if nr >= (bookmarksPerPage*oldalSzam) && nr < (bookmarksPerPage*(oldalSzam+1)) {
			p := template.New("bookmark")
			p2 := template.New("tags")
			tmp := i
			time1, _ := time.Parse("2006-01-02 15:04", tmp.CDate)
			time2, _ := time.Parse("2006-01-02 15:04", tmp.MDate)
			tmp.CDate = timeAgo(time1)
			tmp.MDate = timeAgo(time2)
			var display_bookmark string
			if online {
				display_bookmark = fmt.Sprintf("\n\n\t<div class=\"flip-container\">\n\t\t<div class=\"flipper\" onclick=\"this.classList.toggle('flipped')\">\n\t\t\t<div class=\"front\">\n\t\t\t\t<div class=\"front-side\">\n\t\t\t\t\t<a href=\"%s\">\n\t\t\t\t\t\t<img class=\"honlap_kep_front\" src=\"http://localhost:8080/Required/Images/Screenshots/screenshot_%d.jpg\" alt=\"Website screenshot\">\n\t\t\t\t\t</a>\n\t\t\t\t\t<div class=\"leiras\">\n\t\t\t\t\t\t<a class=\"flip_cim\" href=\"%s\">%s</a>\n\t\t\t\t\t\t<p class=\"flip_description\">%s</p>", tmp.Link, tmp.ID, tmp.Link, tmp.Title, tmp.Description)
			} else {
				display_bookmark = fmt.Sprintf("\n\n\t<div class=\"flip-container\">\n\t\t<div class=\"flipper\" onclick=\"this.classList.toggle('flipped')\">\n\t\t\t<div class=\"front\">\n\t\t\t\t<div class=\"front-side\">\n\t\t\t\t\t<a href=\"http://localhost:8080/Required/Websites/%d/%d.html\">\n\t\t\t\t\t\t<img class=\"honlap_kep_front\" src=\"http://localhost:8080/Required/Images/Screenshots/screenshot_%d.jpg\" alt=\"Website screenshot\">\n\t\t\t\t\t</a>\n\t\t\t\t\t<div class=\"leiras\">\n\t\t\t\t\t\t<a class=\"flip_cim\" href=\"http://localhost:8080/Required/Websites/%d/%d.html\">%s - offline</a>\n\t\t\t\t\t\t<p class=\"flip_description\">%s</p>", tmp.ID, tmp.ID, tmp.ID, tmp.ID, tmp.ID, tmp.Title, tmp.Description)					
			}
			bookmarkTags := "\n\t\t\t\t\t\t<div class=\"flip_all_tags_front\">"
			bookmarkTags2 := "\n\t\t\t\t\t\t<div class=\"flip_all_tags_back\">"
			for _, tags := range strings.Split(tmp.Tags, ";") {
				bookmarkTags += fmt.Sprintf("\n\t\t\t\t\t\t\t<a class=\"flip_tag\" href=\"/bookmarks/%s/page=0/tag=%s\"><strong>%s</strong></a>", strings.Split(InUser.Email, "@")[0], tags, tags)
				bookmarkTags2 += fmt.Sprintf("\n\t\t\t\t\t\t\t<a class=\"flip_tag\" href=\"/bookmarks/%s/page=0/tag=%s\"><strong>%s</strong></a>", strings.Split(InUser.Email, "@")[0], tags, tags)
			}

			bookmarkTags += fmt.Sprintf("\n\t\t\t\t\t\t</div>\n\t\t\t\t\t</div>\n\t\t\t\t</div>\n\t\t\t</div>")
			if online {
				bookmarkTags += fmt.Sprintf("\n\t\t\t<div class=\"back-side\">\n\t\t\t\t<a href=\"%s\">\n\t\t\t\t\t<img class=\"honlap_kep_front\" src=\"http://localhost:8080/Required/Images/Screenshots/screenshot_%d.jpg\" alt=\"Website screenshot\">\n\t\t\t\t</a>\n\t\t\t\t<div class=\"leiras\">\n\t\t\t\t\t<a class=\"flip_cim\" href=\"%s\">%s</a>", tmp.Link, tmp.ID, tmp.Link, tmp.Title)
			} else {
				bookmarkTags += fmt.Sprintf("\n\t\t\t<div class=\"back-side\">\n\t\t\t\t<a href=\"http://localhost:8080/Required/Websites/%d/%d.html\">\n\t\t\t\t\t<img class=\"honlap_kep_front\" src=\"http://localhost:8080/Required/Images/Screenshots/screenshot_%d.jpg\" alt=\"Website screenshot\">\n\t\t\t\t</a>\n\t\t\t\t<div class=\"leiras\">\n\t\t\t\t\t<a class=\"flip_cim\" href=\"http://localhost:8080/Required/Websites/%d/%d.html\">%s - offline</a>", tmp.ID, tmp.ID, tmp.ID, tmp.ID, tmp.ID, tmp.Title)					
			}
			bookmarkTags += bookmarkTags2 + "\n\t\t\t\t\t</div>"
			bookmarkTags += fmt.Sprintf("\n\t\t\t\t\t<p class=\"flip_creationdate\">Created on: %s<br>Last modified on: %s</p>", tmp.CDate, tmp.MDate)
			bookmarkTags += fmt.Sprintf("\n\t\t\t\t\t<a href='/edit_bookmark/bookmarkID=%d'><img class=\"flip_edit\" src=\"http://localhost:8080/Required/Images/Buttons/edit.png\" alt=\"Edit bookmark\"></a>", tmp.ID)
			bookmarkTags += fmt.Sprintf("\n\t\t\t\t\t<a href='/delete_bookmark/bookmarkID=%d'><img class=\"flip_delete\" src=\"http://localhost:8080/Required/Images/Buttons/delete.png\" alt=\"Delete bookmark\"></a>", tmp.ID)
			bookmarkTags += "\n\t\t\t\t</div>\n\t\t\t</div>\n\t\t</div>\n\t</div>"
			p.Parse(display_bookmark)
			p.Execute(w, nil)
			p2.Parse(bookmarkTags)
			p2.Execute(w, nil)
		}
	}
	savHTMLalso := "\n\t<div id=\"also_lepteto\">\n" + savHTML
	t2_2.Parse(savHTMLalso)
	t2_2.Execute(w, nil)
	tagCloud += "\n</div>"
	t3.Parse(tagCloud)
	t3.Execute(w, nil)
	t4.Parse("\n</div>\n</body>\n</html>")
	t4.Execute(w, nil)
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	oldalSzam, _ := strconv.Atoi(vars["page"])
	tagCloudNr, _ := strconv.Atoi(vars["tagcloud"])
	search := vars["search"]

	log.Info("Search handler is started, with %s string.", search)
	t := template.New("tag")
	t, _ = template.ParseFiles("Required/HTML templates/user_interface.html")
	t.Execute(w, InUser)
	t2 := template.New("savHTML")
	t2_1 := template.New("savHTML1")
	t2_2 := template.New("savHTML2")
	t3 := template.New("tadCloud")
	t4 := template.New("end")
	savHTML := ""

	allBookmarksList := database.Search(search, InUser.Email)
	var last int
	if len(allBookmarksList) > bookmarksPerPage {
		if oldalSzam != 0 {
			savHTML = savHTML + fmt.Sprintf("\t<a class=\"hvr-bounce-to-top\" href=\"/search/string=%s/page=%d\"><strong>Previous</strong></a> | \n\t",search, oldalSzam-1)
		}

		for i := 0; i < len(allBookmarksList)/bookmarksPerPage; i++ {
			if i != 0 {
				savHTML = savHTML + " | \n\t"
			}
			if i==oldalSzam{
				savHTML = savHTML + fmt.Sprintf("\t<a class=\"hvr-bounce-to-top2\"><strong>%d-%d</strong></a>", (i*bookmarksPerPage)+1, ((i+1)*bookmarksPerPage))
			} else {
				savHTML = savHTML + fmt.Sprintf("\t<a class=\"hvr-bounce-to-top\" href=\"/search/string=%s/page=%d\"><strong>%d-%d</strong></a>", search, i, (i*bookmarksPerPage)+1, ((i+1)*bookmarksPerPage))
			}
			last = i
		}
		if len(allBookmarksList) > bookmarksPerPage && len(allBookmarksList)%bookmarksPerPage != 0 {
			if oldalSzam==(len(allBookmarksList)/bookmarksPerPage) {
				savHTML = savHTML + fmt.Sprintf(" | \n\t\t<a class=\"hvr-bounce-to-top2\"><strong>%d-%d</strong></a>", ((last+1)*bookmarksPerPage)+1, ((last+1)*bookmarksPerPage)+len(allBookmarksList)%bookmarksPerPage)
 			} else {
				savHTML = savHTML + fmt.Sprintf(" | \n\t\t<a class=\"hvr-bounce-to-top\" href=\"/search/string=%s/page=%d\"><strong>%d-%d</strong></a>", search, last+1, ((last+1)*bookmarksPerPage)+1, ((last+1)*bookmarksPerPage)+len(allBookmarksList)%bookmarksPerPage)
			}
		}
		if oldalSzam < len(allBookmarksList)/bookmarksPerPage {
			savHTML = savHTML + fmt.Sprintf(" | \n\t\t<a class=\"hvr-bounce-to-top\" href=\"/search/string=%s/page=%d\"><strong>Next</strong></a>", search, oldalSzam+1)
		}
	}

	tagCloud := fmt.Sprintf(`	
	</div>
</div>
		
<div class="right-column">
	<div class="tagcloud_header_box">
		<a class="tagcloud_header" href="/search/string=%s/page=%d/tagcloud=0">Tag Cloud</a>
		<a class="tagcloud_header_numbers" href="/search/string=%s/page=%d/tagcloud=20">20</a>
		<a class="tagcloud_header_numbers" href="/search/string=%s/page=%d/tagcloud=10">10</a>
		<a class="tagcloud_header_numbers" href="/search/string=%s/page=%d/tagcloud=5">5</a>
		<a class="tagcloud_header_numbers" href="/search/string=%s/page=%d/tagcloud=2">2</a>
	</div>
	<br>`, search, oldalSzam, search, oldalSzam, search, oldalSzam, search, oldalSzam, search, oldalSzam)
	tagCloudMap := database.TagsCount(InUser.Email, tagCloudNr)
	var tagCloudKeys []string
	for k := range tagCloudMap {
		tagCloudKeys = append(tagCloudKeys, k)
	}
	sort.Strings(tagCloudKeys)
	for _, k := range tagCloudKeys {
		var size int
		if tagCloudMaxSize < 15+3*tagCloudMap[k] {
			size = tagCloudMaxSize
		} else {
			size = 15 + 3*tagCloudMap[k]
		}
		tagCloud += fmt.Sprintf(`
	<a class= "hvr-bob" href="/bookmarks/%s/page=0/tag=%s/tagcloud=0" style="color:#aa5511;font-size:%dpx;">%s </a>`,
			InUser.Email, k, size, k)
	}

	savHTML = savHTML + "\n\t</div>\n"
		t2_1.Parse(`<div id="alatta-wrap">
	<div class="left-column">
	<div id="felso_lepteto">
	`)
	t2_1.Execute(w, nil)
	t2.Parse(savHTML)
	t2.Execute(w, nil)

	var online bool
		if isOnline() {
			online = true
		} else {
			online = false
		}
	for nr, i := range allBookmarksList {
		if nr >= (bookmarksPerPage*oldalSzam) && nr < (bookmarksPerPage*(oldalSzam+1)) {
			p := template.New("bookmark")
			p2 := template.New("tags")
			tmp := i
			time1, _ := time.Parse("2006-01-02 15:04", tmp.CDate)
			time2, _ := time.Parse("2006-01-02 15:04", tmp.MDate)
			tmp.CDate = timeAgo(time1)
			tmp.MDate = timeAgo(time2)
			var display_bookmark string
			if online {
				display_bookmark = fmt.Sprintf("\n\n\t<div class=\"flip-container\">\n\t\t<div class=\"flipper\" onclick=\"this.classList.toggle('flipped')\">\n\t\t\t<div class=\"front\">\n\t\t\t\t<div class=\"front-side\">\n\t\t\t\t\t<a href=\"%s\">\n\t\t\t\t\t\t<img class=\"honlap_kep_front\" src=\"http://localhost:8080/Required/Images/Screenshots/screenshot_%d.jpg\" alt=\"Website screenshot\">\n\t\t\t\t\t</a>\n\t\t\t\t\t<div class=\"leiras\">\n\t\t\t\t\t\t<a class=\"flip_cim\" href=\"%s\">%s</a>\n\t\t\t\t\t\t<p class=\"flip_description\">%s</p>", tmp.Link, tmp.ID, tmp.Link, tmp.Title, tmp.Description)
			} else {
				display_bookmark = fmt.Sprintf("\n\n\t<div class=\"flip-container\">\n\t\t<div class=\"flipper\" onclick=\"this.classList.toggle('flipped')\">\n\t\t\t<div class=\"front\">\n\t\t\t\t<div class=\"front-side\">\n\t\t\t\t\t<a href=\"http://localhost:8080/Required/Websites/%d/%d.html\">\n\t\t\t\t\t\t<img class=\"honlap_kep_front\" src=\"http://localhost:8080/Required/Images/Screenshots/screenshot_%d.jpg\" alt=\"Website screenshot\">\n\t\t\t\t\t</a>\n\t\t\t\t\t<div class=\"leiras\">\n\t\t\t\t\t\t<a class=\"flip_cim\" href=\"http://localhost:8080/Required/Websites/%d/%d.html\">%s - offline</a>\n\t\t\t\t\t\t<p class=\"flip_description\">%s</p>", tmp.ID, tmp.ID, tmp.ID, tmp.ID, tmp.ID, tmp.Title, tmp.Description)					
			}
			bookmarkTags := "\n\t\t\t\t\t\t<div class=\"flip_all_tags_front\">"
			bookmarkTags2 := "\n\t\t\t\t\t\t<div class=\"flip_all_tags_back\">"
			for _, tags := range strings.Split(tmp.Tags, ";") {
				bookmarkTags += fmt.Sprintf("\n\t\t\t\t\t\t\t<a class=\"flip_tag\" href=\"/bookmarks/%s/page=0/tag=%s\"><strong>%s</strong></a>", strings.Split(InUser.Email, "@")[0], tags, tags)
				bookmarkTags2 += fmt.Sprintf("\n\t\t\t\t\t\t\t<a class=\"flip_tag\" href=\"/bookmarks/%s/page=0/tag=%s\"><strong>%s</strong></a>", strings.Split(InUser.Email, "@")[0], tags, tags)
			}

			bookmarkTags += fmt.Sprintf("\n\t\t\t\t\t\t</div>\n\t\t\t\t\t</div>\n\t\t\t\t</div>\n\t\t\t</div>")
			if online {
				bookmarkTags += fmt.Sprintf("\n\t\t\t<div class=\"back-side\">\n\t\t\t\t<a href=\"%s\">\n\t\t\t\t\t<img class=\"honlap_kep_front\" src=\"http://localhost:8080/Required/Images/Screenshots/screenshot_%d.jpg\" alt=\"Website screenshot\">\n\t\t\t\t</a>\n\t\t\t\t<div class=\"leiras\">\n\t\t\t\t\t<a class=\"flip_cim\" href=\"%s\">%s</a>", tmp.Link, tmp.ID, tmp.Link, tmp.Title)
			} else {
				bookmarkTags += fmt.Sprintf("\n\t\t\t<div class=\"back-side\">\n\t\t\t\t<a href=\"http://localhost:8080/Required/Websites/%d/%d.html\">\n\t\t\t\t\t<img class=\"honlap_kep_front\" src=\"http://localhost:8080/Required/Images/Screenshots/screenshot_%d.jpg\" alt=\"Website screenshot\">\n\t\t\t\t</a>\n\t\t\t\t<div class=\"leiras\">\n\t\t\t\t\t<a class=\"flip_cim\" href=\"http://localhost:8080/Required/Websites/%d/%d.html\">%s - offline</a>", tmp.ID, tmp.ID, tmp.ID, tmp.ID, tmp.ID, tmp.Title)					
			}
			bookmarkTags += bookmarkTags2 + "\n\t\t\t\t\t</div>"
			bookmarkTags += fmt.Sprintf("\n\t\t\t\t\t<p class=\"flip_creationdate\">Created on: %s<br>Last modified on: %s</p>", tmp.CDate, tmp.MDate)
			bookmarkTags += fmt.Sprintf("\n\t\t\t\t\t<a href='/edit_bookmark/bookmarkID=%d'><img class=\"flip_edit\" src=\"http://localhost:8080/Required/Images/Buttons/edit.png\" alt=\"Edit bookmark\"></a>", tmp.ID)
			bookmarkTags += fmt.Sprintf("\n\t\t\t\t\t<a href='/delete_bookmark/bookmarkID=%d'><img class=\"flip_delete\" src=\"http://localhost:8080/Required/Images/Buttons/delete.png\" alt=\"Delete bookmark\"></a>", tmp.ID)
			bookmarkTags += "\n\t\t\t\t</div>\n\t\t\t</div>\n\t\t</div>\n\t</div>"
			p.Parse(display_bookmark)
			p.Execute(w, nil)
			p2.Parse(bookmarkTags)
			p2.Execute(w, nil)
		}
	}
	savHTMLalso := "\n\t<div id=\"also_lepteto\">\n" + savHTML
	t2_2.Parse(savHTMLalso)
	t2_2.Execute(w, nil)
	tagCloud += "\n</div>"
	t3.Parse(tagCloud)
	t3.Execute(w, nil)
	t4.Parse("\n</div>\n</body>\n</html>")
	t4.Execute(w, nil)
}

func isOnline() bool{
	log.Info("Determining if there is internet connection function started.")
	out, _:=exec.Command("ping", "www.google.com").Output()
	if(strings.Contains(string(out),"Ping request could not find host") || strings.Contains(string(out), "General failure.") ||
		strings.Contains(string(out),"ping: unknown host")){
		log.Error("No internet connection!")
		return false
	} else {
		log.Info("There is internet connection!")
		return true
	}
}

func timeAgo(t time.Time) string {
	t = t.UTC()
	now := time.Now().UTC().Add(1 * time.Hour)
	diff := now.Sub(t)
	if diff.Minutes() < 1 {
		return "Less than a minute ago"
	} else if diff.Minutes() < 60 {
		return fmt.Sprintf("%.0f minutes ago", diff.Minutes())
	} else if diff.Hours() < 24 {
		return fmt.Sprintf("%.0f hours ago", diff.Hours())
	} else if diff.Hours() < (7 * 24) {
		return fmt.Sprintf("%.0f days ago", (diff.Hours() / 24))
	} else if diff.Hours() < (4 * 7 * 24) {
		return fmt.Sprintf("%.0f weeks ago", (diff.Hours())/(7*24))
	}
	return t.Format("2006-01-02 15:04")
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stderr)
	log.SetLevel(log.WarnLevel)
}
