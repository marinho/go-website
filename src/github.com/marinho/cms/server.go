package cms

import (
    "fmt"
    "log"
    "io"
    "io/ioutil"
    "flag"
    "os"
    "encoding/json"
    "path/filepath"
    "net/http"
    "net/url"
    "strconv"
    "errors"
    "strings"
    "labix.org/v2/mgo"
    "github.com/gorilla/mux"
    "github.com/gorilla/sessions"
    //"github.com/marinho/cms"
)

const VERSION = "0.1"
const HTTP_ADDRESS = ":8080"
const DEFAULT_AUTHOR = "Mario"

var dbDefaultConn *mgo.Session
var sessionStore = sessions.NewCookieStore([]byte("mbSessionId"))

/* Configuration and parameters */

type MenuItem struct {
    Url string
    Id string
    Label string
}

type Configuration struct {
    DBHostname string
    DBName string

    StaticRoot string
    TemplatesRoot string
    AuthSecret string
    AdminUsername string
    AdminPassword string
}
var systemConf Configuration

func defaultConfiguration() Configuration {
    curDir, err := os.Getwd()
    if err != nil {
        curDir = ""
    }
    return Configuration{DBHostname:"localhost", DBName:"mb", StaticRoot:filepath.Join(curDir,"static"),
        TemplatesRoot:filepath.Join(curDir,"templates"), AuthSecret:"", AdminUsername:"admin",
        AdminPassword:"123"}
}

func loadConfiguration(filePath string) Configuration {
    // Read configuration file
    reader, err := os.Open(filePath)
    if err != nil {
        log.Printf("Configuration file %v couldn't be loaded", filePath)
        return defaultConfiguration()
    }

    // Initializing conf instance
    var conf Configuration

    // Parsing JSON from content
    dec := json.NewDecoder(reader)
    if err = dec.Decode(&conf); err == io.EOF {
        log.Printf("Configuration file %v is empty", filePath)
        return defaultConfiguration()
    } else if err != nil {
        log.Printf("Configuration file %v is invalid", filePath)
        return defaultConfiguration()
    }

    return conf
}

func saveConfigFile(conf Configuration, filePath string) {
    b, err := json.Marshal(conf)
    if err == nil {
        err2 := ioutil.WriteFile(filePath, b, os.ModeSetuid | 0750)
        if err2 != nil {
            log.Fatal(err2)
        }
    }
}

type CommandParameters struct {
    Help bool
    ConfigurationFile string
}

func loadParameters() *CommandParameters {
    params := new(CommandParameters)

    // Flags definition
    flag.StringVar(&params.ConfigurationFile, "config", filepath.Join("config/local.json"),
                   "Inform configuration file path")
    flag.BoolVar(&params.Help, "help", false, "Show help information")

    // Flags parsing to load parameters
    flag.Parse()

    return params
}

func showHelp() {
    fmt.Printf("marinhobrandao.com v. %s\n\n", VERSION)
    flag.PrintDefaults()
    fmt.Println("")
}

// Content URL handlers

func renderTemplate(templateName string) (string, error) {
    var err error

    // Base template
    base_content, err := ioutil.ReadFile(filepath.Join(systemConf.TemplatesRoot,"base.html"))
    if err != nil {
        return "", errors.New("Couldn't load base.html")
    } else if templateName == "base.html" {
        return string(base_content), nil
    }

    // Template file
    content, err := ioutil.ReadFile(filepath.Join(systemConf.TemplatesRoot,templateName))
    if err != nil {
        return "", errors.New("Couldn't load " + templateName)
    }

    return strings.Replace(string(base_content), "<!-- CONTENT -->", string(content), 1), nil
}

func renderAdminTemplate(templateName string) (string, error) {
    var err error

    // Base template
    base_content, err := ioutil.ReadFile(filepath.Join(systemConf.TemplatesRoot,"admin","base.html"))
    if err != nil {
        return "", errors.New("Couldn't load admin/base.html")
    } else if templateName == "base.html" {
        return string(base_content), nil
    }

    // Template file
    content, err := ioutil.ReadFile(filepath.Join(systemConf.TemplatesRoot,"admin",templateName))
    if err != nil {
        return "", errors.New("Couldn't load " + templateName)
    }

    return strings.Replace(string(base_content), "<!-- CONTENT -->", string(content), 1), nil
}

func GetSession(c http.ResponseWriter, req *http.Request) (*sessions.Session, error) {
    return sessionStore.Get(req, "mbSession")
}

func IsSuperuserHandler(c http.ResponseWriter, req *http.Request) {
    log.Println(req.URL)
    var data string

    c.Header().Add("Content-Type", "text/plain")

    // Get current session
    session, err := GetSession(c, req)
    if err != nil || session.Values["secret"] != systemConf.AuthSecret {
        data = "no"
    } else {
        data = "yes"
    }

    c.Header().Add("Content-Length", strconv.Itoa(len(data)))
    io.WriteString(c, data)

}

// Decorator for URL handlers whose require superuser authentication
func RequireSuperuser(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
    return func (c http.ResponseWriter, req *http.Request) {
        // Gets the current session
        session, err := GetSession(c, req)
        if err != nil || session.Values["secret"] != systemConf.AuthSecret {
            // Return error
            http.Error(c, "Unauthorized", http.StatusUnauthorized)
            return
        }

        // Call the encapsulated function
        handler(c, req)
    }
}

// Home page handler using template home.html
func HomeHandler(c http.ResponseWriter, req *http.Request) {
    log.Println(req.URL)
    c.Header().Add("Content-Type", "text/html")

    // Not found
    if req.URL.Path != "/" {
        http.Error(c, req.URL.Path + " not found", http.StatusNotFound)
        return
    }

    content, err := renderTemplate("base.html")
    if err != nil {
        log.Println(err)
        c.Header().Add("Content-Length", strconv.Itoa(len("Failed")))
        io.WriteString(c, "Failed")
    } else {
        c.Header().Add("Content-Length", strconv.Itoa(len(content)))
        io.WriteString(c, content)
    }
}

// Login page handler
func LoginHandler(c http.ResponseWriter, req *http.Request) {
    var data string
    var err error
    var session *sessions.Session

    log.Println(req.URL)
    c.Header().Add("Content-Type", "text/json")

    // Validates user
    body, err := ioutil.ReadAll(req.Body)
    if err == nil {
        postValues, err := url.ParseQuery(string(body))
        if err == nil {
            // User validation
            if len(postValues["Username"]) > 0 && postValues["Username"][0] == systemConf.AdminUsername &&
               len(postValues["Password"]) > 0 && postValues["Password"][0] == systemConf.AdminPassword {
                // Starts a session
                session, err = GetSession(c, req)
                if err == nil {
                    session.Values["secret"] = systemConf.AuthSecret
                    session.Save(req, c)
                }
            }
        }
    }

    if session == nil {
        data = fmt.Sprintf("{\"result\":\"error\", \"message\":\"Invalid login\"}")
    } else {
        data = fmt.Sprintf("{\"result\":\"ok\", \"message\":\"User logged successfully\"}")
    }

    c.Header().Add("Content-Length", strconv.Itoa(len(data)))
    io.WriteString(c, data)
}

// Login page handler
func LogoutHandler(c http.ResponseWriter, req *http.Request) {
    // Gets the current session
    session, err := GetSession(c, req)
    if err == nil {
        session.Values["secret"] = "no"
        session.Save(req, c)
    }

    // Redirects to home page
    http.Redirect(c, req, "/admin/", 302)
}

// Menu items handler for the API
func MenuItemsHandler(c http.ResponseWriter, req *http.Request) {
    log.Println(req.URL)
    var data string

    c.Header().Add("Content-Type", "text/json")

    // Menu items list - TODO: move this to database and create a function to load fixtures
    menuItemsList := make([]MenuItem,0)
    menuItemsList = append(menuItemsList, MenuItem{Url:"/", Id:"menu-home", Label:"Home"})
    menuItemsList = append(menuItemsList, MenuItem{Url:"/real-life/", Id:"menu-life", Label:"Real life"})
    menuItemsList = append(menuItemsList, MenuItem{Url:"/legacy/", Id:"menu-legacy", Label:"Legacy"})
    menuItemsList = append(menuItemsList, MenuItem{Url:"http://github.com/marinho", Id:"menu-github", Label:"Github"})
    menuItemsList = append(menuItemsList, MenuItem{Url:"http://old.marinhobrandao.com/", Id:"menu-old", Label:"Old site"})
    menuItemsList = append(menuItemsList, MenuItem{Url:"https://plus.google.com/108430754321695774288/posts", Id:"menu-gplus", Label:"Google+"})
    menuItemsList = append(menuItemsList, MenuItem{Url:"http://de.linkedin.com/in/marinhobrandao", Id:"menu-linkedin", Label:"Linkedin"})

    // Encoding to JSON
    b, err := json.Marshal(menuItemsList)
	if err == nil {
        data = "{\"items\":" + string(b) + "}"
	} else {
		fmt.Println("error:", err)
        data = "{}"
    }

    c.Header().Add("Content-Length", strconv.Itoa(len(data)))
    io.WriteString(c, data)
}

// Blog posts list handler for the API
func BlogPostListHandler(c http.ResponseWriter, req *http.Request) {
    log.Println(req.URL)
    data := "{\"posts\":[]}"

    c.Header().Add("Content-Type", "text/json")

    // Posts from database
    blogPostsList, err := GetRecentBlogPosts(dbDefaultConn.DB(systemConf.DBName))
    if err == nil {
        // Encoding to JSON
        b, err := json.Marshal(blogPostsList)
        if err == nil {
            data = "{\"posts\":" + string(b) + "}"
        } else {
            fmt.Println("error:", err)
        }
    }

    c.Header().Add("Content-Length", strconv.Itoa(len(data)))
    io.WriteString(c, data)
}

// Handler to add a new blog post, for the API
func BlogPostAddHandler(c http.ResponseWriter, req *http.Request) {
    log.Println(req.URL)
    c.Header().Add("Content-Type", "text/json")
    var data string
    var err error
    var blogPost BlogPost
    tags := make([]string,0)

    // Method not allowed
    if req.Method != "POST" {
        http.Error(c, "Invalid method.", http.StatusMethodNotAllowed)
        return
    }

    // Save the new post
    body, err := ioutil.ReadAll(req.Body)
    if err == nil {
        postValues, err := url.ParseQuery(string(body))
        if err == nil {
            if len(postValues["Title"]) == 0 {
                err = errors.New("Title is required")
            } else if len(postValues["Content"]) == 0 {
                err = errors.New("Content is required")
            } else {
                title := postValues["Title"][0]
                content := postValues["Content"][0]
                slug := Slugify(title)
                if len(postValues["Tags"]) > 0 {
                    tags2 := strings.Split(postValues["Tags"][0], ",")
                    for iTag := range tags2 {
                        if strings.Trim(tags2[iTag], " ") != "" {
                            tags = append(tags, strings.Trim(tags2[iTag], " "))
                        }
                    }
                }

                blogPost = BlogPost{Title:title, Content:content, Published:true, Slug:slug, Author:DEFAULT_AUTHOR, Tags:tags}
                err = InsertNewBlogPost(dbDefaultConn.DB(systemConf.DBName), &blogPost)
            }
        }
    }

    if err == nil {
        data = fmt.Sprintf("{\"result\":\"ok\", \"postId\":\"%v\"}", blogPost.Id.Hex())
    } else {
        data = fmt.Sprintf("{\"result\":\"error\"}, \"message\":\"%v\"}", err)
    }

    c.Header().Add("Content-Length", strconv.Itoa(len(data)))
    io.WriteString(c, data)
}

func BlogPostInfoHandler(c http.ResponseWriter, req *http.Request) {
    log.Println(req.URL)
    c.Header().Add("Content-Type", "text/json")
    var post BlogPost
    var err error
    tags := make([]string,0)
    data := "{\"result\":\"error\"}"

    // Parse arguments
    args := mux.Vars(req)

    // Loading blog post
    post, err = GetBlogPost(dbDefaultConn.DB(systemConf.DBName), args["postId"])

    // Renders the template
    if err != nil {
        http.Error(c, "Not found", http.StatusNotFound)
        return
    }

    // Method to return post info
    if req.Method == "GET" {
        // Encoding to JSON
        b, err := json.Marshal(post)
        if err == nil {
            data = "{\"result\":\"ok\", \"post\":" + string(b) + "}"
        } else {
            fmt.Println("error:", err)
        }

    // Method to update post object
    } else if req.Method == "POST" {
        // Gets the current session
        session, err := GetSession(c, req)
        if err != nil || session.Values["secret"] != systemConf.AuthSecret {
            // Return error
            http.Error(c, "Unauthorized", http.StatusUnauthorized)
            return
        }

        body, err := ioutil.ReadAll(req.Body)
        if err == nil {
            postValues, err := url.ParseQuery(string(body))
            if err == nil {
                if len(postValues["Title"]) == 0 {
                    err = errors.New("Title is required")
                } else if len(postValues["Content"]) == 0 {
                    err = errors.New("Content is required")
                } else if len(postValues["Slug"]) == 0 {
                    err = errors.New("Slug is required")
                } else {
                    post.Title = postValues["Title"][0]
                    post.Content = postValues["Content"][0]
                    post.Slug = postValues["Slug"][0]
                    if len(postValues["Tags"]) > 0 {
                        tags2 := strings.Split(postValues["Tags"][0], ",")
                        for iTag := range tags2 {
                            if strings.Trim(tags2[iTag], " ") != "" {
                                tags = append(tags, strings.Trim(tags2[iTag], " "))
                            }
                        }
                    }
                    post.Tags = tags

                    err = UpdateBlogPost(dbDefaultConn.DB(systemConf.DBName), &post)
                }
            }
        }

        // Bad request
        if err != nil {
            http.Error(c, fmt.Sprintf("Bad request: %v", err), http.StatusBadRequest)
            return
        }

    // Method not allowed
    } else {
        http.Error(c, "Invalid method.", http.StatusMethodNotAllowed)
        return
    }

    c.Header().Add("Content-Length", strconv.Itoa(len(data)))
    io.WriteString(c, data)
}

// Handler to delete an existing blog post, for the API
func BlogPostDeleteHandler(c http.ResponseWriter, req *http.Request) {
    log.Println(req.URL)
    c.Header().Add("Content-Type", "text/json")
    var data string
    var err error
    var blogPost BlogPost

    // Method not allowed
    if req.Method != "POST" {
        http.Error(c, "Invalid method.", http.StatusMethodNotAllowed)
        return
    }

    // Parse arguments
    args := mux.Vars(req)

    // Load blog post
    err = DeleteBlogPost(dbDefaultConn.DB(systemConf.DBName), args["postId"])

    if err != nil {
        http.Error(c, "Not found", http.StatusNotFound)
        return
    }

    if err == nil {
        data = fmt.Sprintf("{\"result\":\"ok\", \"postId\":\"%v\"}", blogPost.Id.Hex())
    } else {
        data = fmt.Sprintf("{\"result\":\"error\"}, \"message\":\"%v\"}", err)
    }

    c.Header().Add("Content-Length", strconv.Itoa(len(data)))
    io.WriteString(c, data)
}

// Pages list handler for the API
func PageListHandler(c http.ResponseWriter, req *http.Request) {
    log.Println(req.URL)
    data := "{\"pages\":[]}"

    c.Header().Add("Content-Type", "text/json")

    // Posts from database
    pagesList, err := ListPages(dbDefaultConn.DB(systemConf.DBName))
    if err == nil {
        // Encoding to JSON
        b, err := json.Marshal(pagesList)
        if err == nil {
            data = "{\"pages\":" + string(b) + "}"
        } else {
            fmt.Println("error:", err)
        }
    }

    c.Header().Add("Content-Length", strconv.Itoa(len(data)))
    io.WriteString(c, data)
}

func PageInfoHandler(c http.ResponseWriter, req *http.Request) {
    log.Println(req.URL)
    c.Header().Add("Content-Type", "text/json")
    var page Page
    var err error
    tags := make([]string,0)
    data := "{\"result\":\"error\"}"

    // Parse arguments
    args := mux.Vars(req)

    // Loading page
    if args["pageSlug"] != "" {
        page, err = GetPageBySlug(dbDefaultConn.DB(systemConf.DBName), args["pageSlug"])
    } else {
        page, err = GetPage(dbDefaultConn.DB(systemConf.DBName), args["pageId"])
    }

    // Renders the template
    if err != nil {
        http.Error(c, "Not found", http.StatusNotFound)
        return
    }

    // Method to return page info
    if req.Method == "GET" {
        // Encoding to JSON
        b, err := json.Marshal(page)
        if err == nil {
            data = "{\"result\":\"ok\", \"page\":" + string(b) + "}"
        } else {
            fmt.Println("error:", err)
        }

    // Method to update page object
    } else if req.Method == "POST" {
        // Gets the current session
        session, err := GetSession(c, req)
        if err != nil || session.Values["secret"] != systemConf.AuthSecret {
            // Return error
            http.Error(c, "Unauthorized", http.StatusUnauthorized)
            return
        }

        body, err := ioutil.ReadAll(req.Body)
        if err == nil {
            postValues, err := url.ParseQuery(string(body))
            if err == nil {
                if len(postValues["Title"]) == 0 {
                    err = errors.New("Title is required")
                } else if len(postValues["Content"]) == 0 {
                    err = errors.New("Content is required")
                } else if len(postValues["Slug"]) == 0 {
                    err = errors.New("Slug is required")
                } else {
                    page.Title = postValues["Title"][0]
                    page.Content = postValues["Content"][0]
                    page.Slug = postValues["Slug"][0]
                    if len(postValues["Tags"]) > 0 {
                        tags2 := strings.Split(postValues["Tags"][0], ",")
                        for iTag := range tags2 {
                            if strings.Trim(tags2[iTag], " ") != "" {
                                tags = append(tags, strings.Trim(tags2[iTag], " "))
                            }
                        }
                    }
                    page.Tags = tags

                    err = UpdatePage(dbDefaultConn.DB(systemConf.DBName), &page)
                }
            }
        }

        // Bad request
        if err != nil {
            http.Error(c, fmt.Sprintf("Bad request: %v", err), http.StatusBadRequest)
            return
        }

    // Method not allowed
    } else {
        http.Error(c, "Invalid method.", http.StatusMethodNotAllowed)
        return
    }

    c.Header().Add("Content-Length", strconv.Itoa(len(data)))
    io.WriteString(c, data)
}

// Page presentation -- TODO: remove this and use HomeHandler instead
func PageViewHandler(c http.ResponseWriter, req *http.Request) {
    log.Println(req.URL)
    c.Header().Add("Content-Type", "text/html")
    var data string
    var found bool

    // Method not allowed
    if req.Method != "GET" {
        http.Error(c, "Invalid method.", http.StatusMethodNotAllowed)
        return
    }

    // Parse arguments
    args := mux.Vars(req)

    // Loading page
    if args["pageSlug"] == "404" {
        found = true
    } else {
        found = PageExists(dbDefaultConn.DB(systemConf.DBName), args["pageSlug"])
    }

    // Page not found
    if !found {
        http.Error(c, fmt.Sprintf("Page \"%v\" not found", args["pageSlug"]), http.StatusNotFound)
        return
    }

    // Renders the template
    data, err := renderTemplate("base.html")

    if err != nil {
        http.Error(c, "Server error", http.StatusInternalServerError)
        return
    }

    c.Header().Add("Content-Length", strconv.Itoa(len(data)))
    io.WriteString(c, data)
}

// Handler to add a new page, for the API
func PageAddHandler(c http.ResponseWriter, req *http.Request) {
    log.Println(req.URL)
    c.Header().Add("Content-Type", "text/json")
    var data string
    var err error
    var page Page
    tags := make([]string,0)

    // Method not allowed
    if req.Method != "POST" {
        http.Error(c, "Invalid method.", http.StatusMethodNotAllowed)
        return
    }

    // Save the new post
    body, err := ioutil.ReadAll(req.Body)
    if err == nil {
        postValues, err := url.ParseQuery(string(body))
        if err == nil {
            if len(postValues["Title"]) == 0 {
                err = errors.New("Title is required")
            } else if len(postValues["Content"]) == 0 {
                err = errors.New("Content is required")
            } else if len(postValues["Slug"]) == 0 {
                err = errors.New("Slug is required")
            } else {
                title := postValues["Title"][0]
                content := postValues["Content"][0]
                slug := postValues["Slug"][0]
                if len(postValues["Tags"]) > 0 {
                    tags2 := strings.Split(postValues["Tags"][0], ",")
                    for iTag := range tags2 {
                        if strings.Trim(tags2[iTag], " ") != "" {
                            tags = append(tags, strings.Trim(tags2[iTag], " "))
                        }
                    }
                }

                page = Page{Title:title, Content:content, Published:true, Slug:slug, Author:DEFAULT_AUTHOR, Tags:tags}
                err = InsertNewPage(dbDefaultConn.DB(systemConf.DBName), &page)
            }
        }
    }

    if err == nil {
        data = fmt.Sprintf("{\"result\":\"ok\", \"postId\":\"%v\"}", page.Id.Hex())
    } else {
        data = fmt.Sprintf("{\"result\":\"error\"}, \"message\":\"%v\"}", err)
    }

    c.Header().Add("Content-Length", strconv.Itoa(len(data)))
    io.WriteString(c, data)
}

// Handler to delete an existing page, for the API
func PageDeleteHandler(c http.ResponseWriter, req *http.Request) {
    log.Println(req.URL)
    c.Header().Add("Content-Type", "text/json")
    var data string
    var err error
    var page Page

    // Method not allowed
    if req.Method != "POST" {
        http.Error(c, "Invalid method.", http.StatusMethodNotAllowed)
        return
    }

    // Parse arguments
    args := mux.Vars(req)

    // Load blog post
    err = DeletePage(dbDefaultConn.DB(systemConf.DBName), args["pageId"])

    if err != nil {
        http.Error(c, "Not found", http.StatusNotFound)
        return
    }

    if err == nil {
        data = fmt.Sprintf("{\"result\":\"ok\", \"postId\":\"%v\"}", page.Id.Hex())
    } else {
        data = fmt.Sprintf("{\"result\":\"error\"}, \"message\":\"%v\"}", err)
    }

    c.Header().Add("Content-Length", strconv.Itoa(len(data)))
    io.WriteString(c, data)
}

/* Admin handlers */

// Home page handler for Administration area
func AdminHomeHandler(c http.ResponseWriter, req *http.Request) {
    log.Println(req.URL)

    c.Header().Add("Content-Type", "text/html")

    content, err := renderAdminTemplate("base.html")
    if err != nil {
        log.Println(err)
        c.Header().Add("Content-Length", strconv.Itoa(len("Failed")))
        io.WriteString(c, "Failed")
    } else {
        c.Header().Add("Content-Length", strconv.Itoa(len(content)))
        io.WriteString(c, content)
    }
}

// Menu items handler for the API
func AdminMenuHandler(c http.ResponseWriter, req *http.Request) {
    log.Println(req.URL)
    var data string

    c.Header().Add("Content-Type", "text/json")

    // Menu items list - TODO: move this to database and create a function to load fixtures
    menuItemsList := make([]MenuItem,0)
    menuItemsList = append(menuItemsList, MenuItem{Url:"/admin/", Id:"admin-home", Label:"Home"})
    menuItemsList = append(menuItemsList, MenuItem{Url:"/admin/pages/", Id:"admin-pages", Label:"Pages"})
    menuItemsList = append(menuItemsList, MenuItem{Url:"/admin/blog-posts/", Id:"admin-blog-posts", Label:"Blog Posts"})
    menuItemsList = append(menuItemsList, MenuItem{Url:"/logout/", Id:"admin-logout", Label:"Logout"})

    // Encoding to JSON
    b, err := json.Marshal(menuItemsList)
	if err == nil {
        data = "{\"items\":" + string(b) + "}"
	} else {
		fmt.Println("error:", err)
        data = "{}"
    }

    c.Header().Add("Content-Length", strconv.Itoa(len(data)))
    io.WriteString(c, data)
}

func SetUrls() {
    // URL routes
    r := mux.NewRouter()

    r.HandleFunc("/", HomeHandler)
    r.HandleFunc("/login/", LoginHandler)
    r.HandleFunc("/logout/", LogoutHandler)

    // Admin
    r.HandleFunc("/admin/", AdminHomeHandler)
    r.HandleFunc("/admin/pages/", RequireSuperuser(AdminHomeHandler))
    r.HandleFunc("/admin/blog-posts/", RequireSuperuser(AdminHomeHandler))
    r.HandleFunc("/api/admin/menu/", RequireSuperuser(AdminMenuHandler))

    // General API methods
    r.HandleFunc("/api/is-superuser/", IsSuperuserHandler)
    r.HandleFunc("/api/menu/item/", MenuItemsHandler)

    // Blog posts
    r.HandleFunc("/api/blog/post/", BlogPostListHandler)
    r.HandleFunc("/api/blog/post/{postId:\\w+}/", BlogPostInfoHandler)
    r.HandleFunc("/api/blog/post/add/", RequireSuperuser(BlogPostAddHandler))
    r.HandleFunc("/api/blog/post/{postId:\\w+}/delete/", RequireSuperuser(BlogPostDeleteHandler))

    // Pages
    r.HandleFunc("/api/page/", PageListHandler)
    r.HandleFunc("/api/page/{pageId:[\\w\\-]+}/", PageInfoHandler)
    r.HandleFunc("/api/page/add/", RequireSuperuser(PageAddHandler))
    r.HandleFunc("/api/page/{pageId:\\w+}/delete/", RequireSuperuser(PageDeleteHandler))
    r.HandleFunc("/api/page/by-slug/{pageSlug:[\\w\\-]+}/", PageInfoHandler)
    r.HandleFunc("/{pageSlug:[\\w\\-]+}", PageViewHandler)
    r.HandleFunc("/{pageSlug:[\\w\\-]+}/", PageViewHandler) // This is needed to support both, but maybe there's an alternative

    // Hardcoded ones
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(systemConf.StaticRoot))))
    http.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir(systemConf.TemplatesRoot))))
    http.Handle("/", r)
}

// Main routine
func ServerMain() {
    // Parsing command line parameters
    params := loadParameters()
    var err error

    if params.Help {
        showHelp()
        os.Exit(0)
    }

    // Reading configuration file
    systemConf = loadConfiguration(params.ConfigurationFile)

    // Load connections
    dbDefaultConn, err = mgo.Dial(systemConf.DBHostname)
    if err != nil {
        log.Fatal(err)
    }
    defer dbDefaultConn.Close()

    // Optional. Switch the session to a monotonic behavior.
    dbDefaultConn.SetMode(mgo.Monotonic, true)

    SetUrls()

    // Start serving!
    log.Fatal(http.ListenAndServe(HTTP_ADDRESS, nil))
}

