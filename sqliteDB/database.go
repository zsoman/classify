package database

import (
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"fmt"
	"os"
	"io"
	"strconv"
	"strings"
	"time"
	"os/exec"

	log "github.com/Sirupsen/logrus"
	_ "github.com/mxk/go-sqlite/sqlite3"	//Windows
	_ "github.com/mattn/go-sqlite3"	//Linux
)

type Bookmark struct {
	ID          int
	Title       string
	Link        string
	Tags        string
	Categories  string
	Description string
	CDate       string
	MDate       string
	User        string
}

var key string

func IsEmptyUsers() bool {
	log.Info("Is empty useres function started from database package.")
	db, _ := sql.Open("sqlite3", `Required\database.db`)
	rows, err := db.Query("SELECT count(*) FROM users")
	if err != nil {
		db.Exec(`CREATE TABLE bookmarks(
   id	INTEGER	NOT NULL	PRIMARY KEY	AUTOINCREMENT	UNIQUE,
   link	TEXT	NOT NULL,
   tags	TEXT	DEFAULT 'none',
   category	TEXT	DEFAULT 'none',
   cDate	INTEGER NOT NULL,
   mDate	INTEGER NOT NULL,
   title	TEXT	DEFAULT 'Title',
   description	TEXT	DEFAULT 'Description',
   user	TEXT	NOT NULL
)`)
		db.Exec(`CREATE TABLE users(
   id	INTEGER	NOT NULL	PRIMARY KEY	AUTOINCREMENT	UNIQUE,
   email	TEXT	NOT NULL UNIQUE,
   pwd	TEXT	NOT NULL,
   fname	TEXT,
   lname	TEXT
)`)
		log.Error("Because the database.db database was not found in the Required folder an empty database was created.")
		db.Close()
		return false
	}
	defer rows.Close()
	rows.Next()
	var count int
	_ = rows.Scan(&count)
	db.Close()
	if count > 0 {
		return true
	} else {
		log.Warn("No registered users, please register!")
		return false
	}
}

func Login(email, pwd string) bool {
	log.Info("Login function started from database package.")
	var count int
	db, _ := sql.Open("sqlite3", `Required\database.db`)

	email = Encrypt(email)
	pwd = Encrypt(pwd)

	err := db.QueryRow("select count(*) from users where email=? and pwd=?", email, pwd).Scan(&count)
	switch {
	case err == sql.ErrNoRows:
		log.Error(fmt.Sprintf("Error in logging in the %s user.", Decrypt(email)))
	case err != nil:
		log.Fatal(err)
	default:
		if count == 1 {
			log.Warn(fmt.Sprintf("%s user successfully logged in.", Decrypt(email)))
			db.Close()
			DownloadScreenshots()
			return true
		} else {
			if count==0 {
				log.Warn(fmt.Sprintf("No users found with %s username.", Decrypt(email)))
			} else if count>1{
				log.Error(fmt.Sprintf("Multiple users found with %s username.", Decrypt(email)))
			}
		}
	}
	db.Close()
	return false
}

func isOnline() bool{
	log.Info("Determining if there is internet connection function started from database package.")
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

func DownloadScreenshots(){
	log.Info("Download all undownloaded screenshots function started from database package.")
	out, err:=exec.Command("cmd","/C","dir", `Required\Images\Screenshots`).Output()
	db, _ := sql.Open("sqlite3", `Required\database.db`)
	var allScrenshotsID []int
	var alldbID []int
	for nr,i:=range strings.Split(string(out),`screenshot_`){
		if nr!=0 {
			szam:=i[:strings.Index(i,".")]
			szamuj,_:=strconv.Atoi(szam)
			allScrenshotsID = append(allScrenshotsID,szamuj)
		}
	}
    rows, err := db.Query("SELECT id FROM bookmarks")
    checkErr(err)

    for rows.Next() {
        var id int
        rows.Scan(&id)
		alldbID = append(alldbID,id)
    }
	for _,i:=range alldbID{
		bennevan:=false
		for _,j:=range allScrenshotsID{
			if i==j{
				bennevan=true
			}
		}
		if bennevan == false{
			var link string
			db.QueryRow("SELECT link FROM bookmarks WHERE id=?", i).Scan(&link)
			exec.Command(`Required\3rd party software\PhantomJS\bin\phantomjs`, `Required\3rd party software\rasterize-nojs.js`, Decrypt(link), fmt.Sprintf(`Required\Images\Screenshots\screenshot_%d.jpg`,i),"1024px*1228px").Run()
			log.Warn(fmt.Sprintf("screenshot_%d.jpg is downloaded, because thare was no internet connection last time.",i))
		}
	}
	db.Close()
	checkErr(err)
}

func Register(email, pwd, fname, lname string) bool {
	log.Info("User register function started from database package.")
	var count int
	db, err := sql.Open("sqlite3", `Required\database.db`)
	checkErr(err)

	email = Encrypt(email)
	pwd = Encrypt(pwd)
	fname = Encrypt(fname)
	lname = Encrypt(lname)

	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email=?", email).Scan(&count)
	if count > 0 {
		log.Error(fmt.Sprintf("The %s email address is already been registered.",Decrypt(email)))
		return false
	}

	_, err = db.Exec("INSERT INTO users(email,pwd,fname,lname) VALUES(?,?,?,?)", email, pwd, fname, lname)
	db.Close()
	checkErr(err)
	log.Warn(fmt.Sprintf("The new user, with %s email address, is successfully has been registered.",Decrypt(email)))
	return true
}

func NewBookmark(title, link, tags, category, description, cdate, mdate, user string) bool {
	log.Info("New bookmark register function started from database package.")
	db, err := sql.Open("sqlite3", `Required\database.db`)
	var newBookmarkID int
	checkErr(err)
	cdateInt, _ := strconv.ParseInt(cdate, 10, 64)
	mdateInt, _ := strconv.ParseInt(mdate, 10, 64)

	title = Encrypt(title)
	link = Encrypt(link)
	tags = Encrypt(tags)
	category = Encrypt(category)
	description = Encrypt(description)
	user = Encrypt(user)

	_, err = db.Exec("INSERT INTO bookmarks(title, link, tags, category, description, cdate, mdate, user) VALUES(?,?,?,?,?,?,?,?)", title, link, tags, category, description, cdateInt, mdateInt, user)
	log.Warn("The new bookmark has been saved.")
	db.QueryRow("SELECT MAX(id) FROM bookmarks").Scan(&newBookmarkID)
	if(isOnline()){
		exec.Command(`Required\3rd party software\PhantomJS\bin\phantomjs`, `Required\3rd party software\rasterize-nojs.js`, Decrypt(link), fmt.Sprintf(`Required\Images\Screenshots\screenshot_%d.jpg`,newBookmarkID),"1024px*1228px").Run()
		log.Warn(fmt.Sprintf("screenshot_%d.jpg is downloaded.",newBookmarkID))
		f, _ := os.Create(".downloadHTML.bat")
		io.WriteString(f, fmt.Sprintf(`Required\"3rd party software"\WinHTTrack\httrack --get %s -O Required\Websites\%d -N %d.html`,Decrypt(link),newBookmarkID,newBookmarkID))
		f.Close()
		log.Info(".downloadHTML.bat batch file is created and the command is written into it.")
		exec.Command(`.downloadHTML.bat`).Run()
		log.Warn(fmt.Sprintf("The offline version of the website is downloaded into %d folder.",newBookmarkID))
		exec.Command("cmd", "/C", "del", ".downloadHTML.bat").Run()
		log.Info(".downloadHTML.bat batch file is deleted.")
	}
	db.Close()
	return true
}

func EditBookmark(id, cdate int, title, link, tags, category, description, mdate, user string) bool {
	log.Info("Edit bookmark register function started  from database package.")
	db, err := sql.Open("sqlite3", `Required\database.db`)
	checkErr(err)
	mdateInt, _ := strconv.ParseInt(mdate, 10, 64)

	title = Encrypt(title)
	link = Encrypt(link)
	tags = Encrypt(tags)
	category = Encrypt(category)
	description = Encrypt(description)
	user = Encrypt(user)

	stmt, err := db.Prepare("UPDATE bookmarks SET title=?, link=?, tags=?, category=?, description=?, cdate=?, mdate=?, user=? WHERE id=?")
	checkErr(err)
	_, err = stmt.Exec(title, link, tags, category, description, cdate, mdateInt, user, id)
	checkErr(err)
	log.Info("New data is saved in the database for the bookmark that is edited.")

	if(isOnline()){
		err = exec.Command(`Required\3rd party software\PhantomJS\bin\phantomjs`, `Required\3rd party software\rasterize-nojs.js`, Decrypt(link), fmt.Sprintf(`Required\Images\Screenshots\screenshot_%d.jpg`,id),"1024px*1228px").Run()
		log.Warn(fmt.Sprintf("The screenshot_%d.jpg is redownloaded.",id))
		f, _ := os.Create(".downloadHTML.bat")
		io.WriteString(f, fmt.Sprintf(`Required\"3rd party software"\WinHTTrack\httrack --get %s -O Required\Websites\%d -N %d.html`,Decrypt(link),id,id))
		f.Close()
		log.Info(".downloadHTML.bat batch file is created and the command is written into it.")
		exec.Command(`.downloadHTML.bat`).Run()
		log.Warn(fmt.Sprintf("The offline version of the website is redownloaded into %d folder.",id))
		exec.Command("cmd", "/C", "del", ".downloadHTML.bat").Run()
		log.Info(".downloadHTML.bat batch file is deleted.")
	}
	db.Close()
	return true
}

func AllBookmarks(user string) []Bookmark {
	log.Info("Collecting all bookmarks function started from database package.")
	db, err := sql.Open("sqlite3", `Required\database.db`)
	checkErr(err)
	user = Encrypt(user)
	var bookmarksList []Bookmark
	rows, _ := db.Query("SELECT id, title, link, description, tags, category, cdate, mdate FROM bookmarks WHERE user = ? ORDER BY mDate DESC", user)
	log.Info(fmt.Sprintf("All the bookmarks saved by the %s user is collected from the database.",Decrypt(user)))
	for rows.Next() {
		var bookmark Bookmark
		_ = rows.Scan(&bookmark.ID, &bookmark.Title, &bookmark.Link, &bookmark.Description, &bookmark.Tags,
			&bookmark.Categories, &bookmark.CDate, &bookmark.MDate)
		bookmark.Title = Decrypt(bookmark.Title)
		bookmark.Description = Decrypt(bookmark.Description)
		bookmark.Tags = Decrypt(bookmark.Tags)
		bookmark.Categories = Decrypt(bookmark.Categories)
		tmp, _ := strconv.ParseInt(bookmark.CDate, 10, 64)
		bookmark.CDate = time.Unix(tmp, 0).Format("2006-01-02 15:04")
		tmp, _ = strconv.ParseInt(bookmark.MDate, 10, 64)
		bookmark.MDate = time.Unix(tmp, 0).Format("2006-01-02 15:04")
		bookmark.Link = Decrypt(bookmark.Link)
		bookmarksList = append(bookmarksList, bookmark)
	}
	db.Close()
	log.Info("A list of structures in which the bookmarks are stored is created and returned.")
	return bookmarksList
}

func RetrieveOneBookmark(id int) Bookmark {
	log.Info("Retrieve one bookmark function started from database package.")
	db, err := sql.Open("sqlite3", `Required\database.db`)
	checkErr(err)
	var bookmark Bookmark
	db.QueryRow("SELECT id, title, link, description, tags, category, cDate, mDate, user FROM bookmarks WHERE id = ?", id).Scan(&bookmark.ID, &bookmark.Title, &bookmark.Link, &bookmark.Description, &bookmark.Tags, &bookmark.Categories, &bookmark.CDate, &bookmark.MDate, &bookmark.User)
	log.Info("The bookmark is retrieved from the database.")
	bookmark.Title = Decrypt(bookmark.Title)
	bookmark.Description = Decrypt(bookmark.Description)
	bookmark.Tags = Decrypt(bookmark.Tags)
	bookmark.Categories = Decrypt(bookmark.Categories)
	tmp, _ := strconv.ParseInt(bookmark.MDate, 10, 64)
	bookmark.MDate = time.Unix(tmp, 0).Format("2006-01-02 15:04")
	bookmark.Link = Decrypt(bookmark.Link)
	bookmark.User = Decrypt(bookmark.User)
	db.Close()
	log.Info("A structure in which the bookmarks are stored is created and returned.")
	return bookmark
}

func DeleteBookmark(id int) {
	log.Info("Delete bookmark function started from database package.")
	db, err := sql.Open("sqlite3", `Required\database.db`)
	checkErr(err)
	stmt, err := db.Prepare("DELETE FROM bookmarks WHERE id = ?")
	checkErr(err)
	stmt.Exec(id)
	checkErr(err)
	log.Info(fmt.Sprintf("The bookmark with %d ID is deleted from the databese.",id))
	exec.Command("cmd", "/C", "del", fmt.Sprintf(`Required\Images\Screenshots\screenshot_%d.jpg`,id)).Run()
	log.Warn(fmt.Sprintf("The screenshot_%d.jpg is deleted.",id))
	exec.Command("cmd", "/C", "rmdir", "/S", "/Q", fmt.Sprintf(`Required\Websites\%d`,id)).Run()
	log.Warn(fmt.Sprintf("The offline version of the bookmark and the %d folder is deleted.",id))
	db.Close()
}

func TagsBookmarks(user, tag string) []Bookmark {
	log.Info("Tags bookmarks function started from database package.")
	db, err := sql.Open("sqlite3", `Required\database.db`)
	checkErr(err)
	user = Encrypt(user)
	var bookmarksList []Bookmark
	rows, _ := db.Query("SELECT id, title, link, description, tags, category, cdate, mdate FROM bookmarks WHERE user = ? ORDER BY mDate DESC", user)
	log.Info(fmt.Sprintf("All of the bookmarks of the %s user is retrieved from the database.",Decrypt(user)))
	for rows.Next() {
		var bookmark Bookmark
		_ = rows.Scan(&bookmark.ID, &bookmark.Title, &bookmark.Link, &bookmark.Description, &bookmark.Tags,
			&bookmark.Categories, &bookmark.CDate, &bookmark.MDate)
		bookmark.Title = Decrypt(bookmark.Title)
		bookmark.Description = Decrypt(bookmark.Description)
		bookmark.Tags = Decrypt(bookmark.Tags)
		bookmark.Categories = Decrypt(bookmark.Categories)
		tmp, _ := strconv.ParseInt(bookmark.CDate, 10, 64)
		bookmark.CDate = time.Unix(tmp, 0).Format("2006-01-02 15:04")
		tmp, _ = strconv.ParseInt(bookmark.MDate, 10, 64)
		bookmark.MDate = time.Unix(tmp, 0).Format("2006-01-02 15:04")
		bookmark.Link = Decrypt(bookmark.Link)
		for _, i := range strings.Split(bookmark.Tags, ";") {
			if i == tag {
				bookmarksList = append(bookmarksList, bookmark)
			}
		}
	}
	db.Close()
	log.Info(fmt.Sprintf("A list of structs in which all of the bookmarks are stored that has a tag: %s.",tag))
	return bookmarksList
}

func Search(str, user string) []Bookmark {
	log.Info("Search function started from database package.")
	db, _ := sql.Open("sqlite3", `Required\database.db`)
	user = Encrypt(user)
	var bookmarksList []Bookmark

	rows, _ := db.Query("SELECT id, title, link, description, tags, category, cdate, mdate FROM bookmarks WHERE user = ? ORDER BY mDate DESC", user)
	log.Info(fmt.Sprintf("All of the bookmarks of the %s user is retrieved from the database.",Decrypt(user)))
	for rows.Next() {
		var bookmark Bookmark
		_ = rows.Scan(&bookmark.ID, &bookmark.Title, &bookmark.Link, &bookmark.Description, &bookmark.Tags,
			&bookmark.Categories, &bookmark.CDate, &bookmark.MDate)
		bookmark.Title = Decrypt(bookmark.Title)
		bookmark.Description = Decrypt(bookmark.Description)
		bookmark.Tags = Decrypt(bookmark.Tags)
		bookmark.Categories = Decrypt(bookmark.Categories)
		tmp, _ := strconv.ParseInt(bookmark.CDate, 10, 64)
		bookmark.CDate = time.Unix(tmp, 0).Format("2006-01-02 15:04")
		tmp, _ = strconv.ParseInt(bookmark.MDate, 10, 64)
		bookmark.MDate = time.Unix(tmp, 0).Format("2006-01-02 15:04")
		bookmark.Link = Decrypt(bookmark.Link)
		if strings.Contains(strings.ToLower(bookmark.Tags), strings.ToLower(str)) {
			bookmarksList = append(bookmarksList, bookmark)
		} else if strings.Contains(strings.ToLower(bookmark.Description), strings.ToLower(str)) {
			bookmarksList = append(bookmarksList, bookmark)
		} else if strings.Contains(strings.ToLower(bookmark.Categories), strings.ToLower(str)) {
			bookmarksList = append(bookmarksList, bookmark)
		} else if strings.Contains(strings.ToLower(bookmark.Title), strings.ToLower(str)) {
			bookmarksList = append(bookmarksList, bookmark)
		} else if strings.Contains(strings.ToLower(bookmark.Link), strings.ToLower(str)) {
			bookmarksList = append(bookmarksList, bookmark)
		}
	}
	db.Close()
	log.Info(fmt.Sprintf("A list of structs in which all of the bookmarks are stored that contains in the title, description, tags or in the link the: %s string.",str))
	return bookmarksList
}

func Encrypt(s string) string {
	log.Info("Encrypting.")
	key = "HabBeMXp0uqHSEM5OnFxDxdkkylEijox"
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(err)
	}
	str := []byte(s)
	ciphertext := []byte("9gEY64fQX7VjLY0WbwtyRqMbBrgOYyEb")
	iv := ciphertext[:aes.BlockSize]
	encrypter := cipher.NewCFBEncrypter(block, iv)
	encrypted := make([]byte, len(str))
	encrypter.XORKeyStream(encrypted, str)
	var tmp string
	for _, i := range encrypted {
		tmp += " " + strconv.Itoa(int(i))
	}
	return tmp[1:]
}

func Decrypt(str string) string {
	log.Info("Decrypting")
	key = "HabBeMXp0uqHSEM5OnFxDxdkkylEijox"
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(err)
	}
	ciphertext := []byte("9gEY64fQX7VjLY0WbwtyRqMbBrgOYyEb")
	iv := ciphertext[:aes.BlockSize]
	decrypter := cipher.NewCFBDecrypter(block, iv) // simple!
	decrypted := make([]byte, len(strings.Split(str, " ")))

	encrypted := make([]byte, len(strings.Split(str, " ")))
	for i, elem := range strings.Split(str, " ") {
		u, _ := strconv.ParseUint(elem, 0, 64)
		encrypted[i] = byte(u)
	}
	decrypter.XORKeyStream(decrypted, encrypted)
	return string(decrypted)
}

func TagsCount(user string, threshold int) map[string]int {
	log.Info("Tags count function started from database package.")
	db, err := sql.Open("sqlite3", `Required\database.db`)
	checkErr(err)
	user = Encrypt(user)
	tags := make(map[string]int)
	tags2 := make(map[string]int)
	rows, _ := db.Query("SELECT tags FROM bookmarks WHERE user = ?", user)
	for rows.Next() {
		var tagRow string
		_ = rows.Scan(&tagRow)
		tagRow = Decrypt(tagRow)
		for _, i := range strings.Split(tagRow, ";") {
			_, ok := tags[i]
			if ok {
				tags[i] = tags[i] + 1
			} else {
				tags[i] = 1
			}
		}
	}
	for key, _ := range tags {
		if tags[key] >= threshold {
			tags2[key] = tags[key]
		}
	}
	db.Close()
	log.Info("A Map with all of the unique tags and with a number that indicates how many times the tag is used is returned.")
	return tags2
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stderr)
	log.SetLevel(log.WarnLevel)
}
