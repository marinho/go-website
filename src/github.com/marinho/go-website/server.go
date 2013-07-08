package main

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
    "time"
    "labix.org/v2/mgo"
    //"labix.org/v2/mgo/bson"
    "github.com/gorilla/mux"
    "github.com/marinho/go-website"
)

const VERSION = "0.1"
const HTTP_ADDRESS = ":8080"
const DEFAULT_AUTHOR = "Mario"

var dbDefaultConn *mgo.Session

/* Configuration and parameters */

type Configuration struct {
    DBHostname string
    DBName string

    StaticRoot string
    TemplatesRoot string
}
var systemConf Configuration

func defaultConfiguration() Configuration {
    curDir, err := os.Getwd()
    if err != nil {
        curDir = ""
    }
    return Configuration{DBHostname:"localhost", DBName:"mb", StaticRoot:filepath.Join(curDir,"static"),
        TemplatesRoot:filepath.Join(curDir,"templates")}
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
    templatePath := filepath.Join(systemConf.TemplatesRoot,templateName)

    content, err := ioutil.ReadFile(templatePath)
    if err != nil {
        return "", errors.New("Couldn't load home.html")
    }

    return string(content), nil
}

func HomeHandler(c http.ResponseWriter, req *http.Request) {
    log.Println(req.URL)

    c.Header().Add("Content-Type", "text/html")

    // Not found
    if req.URL.Path != "/" {
        http.Error(c, req.URL.Path + " not found", http.StatusNotFound)
        return
    }

    content, err := renderTemplate("home.html")
    if err != nil {
        log.Println(err)
        c.Header().Add("Content-Length", strconv.Itoa(len("Failed")))
        io.WriteString(c, "Failed")
    } else {
        c.Header().Add("Content-Length", strconv.Itoa(len(content)))
        io.WriteString(c, content)
    }
}

func LoginHandler(c http.ResponseWriter, req *http.Request) {
    log.Println(req.URL)
    expiration := time.Now().AddDate(1, 0, 0)
    cookie := http.Cookie{Name:"mbAuth", Value:"LetTarLin", Expires:expiration}
    http.SetCookie(c, &cookie)
    http.Redirect(c, req, "/", 302)
    //io.WriteString(c, "hey")
}

type MenuItem struct {
    Url string
    Id string
    Label string
}

func MenuItemsHandler(c http.ResponseWriter, req *http.Request) {
    log.Println(req.URL)
    var data string

    c.Header().Add("Content-Type", "text/json")

    // Menu items list - TODO: move this to database and create a function to load fixtures
    menuItemsList := make([]MenuItem,0)
    menuItemsList = append(menuItemsList, MenuItem{Url:"/", Id:"menu-home", Label:"Home"})
    menuItemsList = append(menuItemsList, MenuItem{Url:"/real-life/", Id:"menu-life", Label:"Real life"})
    menuItemsList = append(menuItemsList, MenuItem{Url:"http://github.com/marinho", Id:"menu-github", Label:"Github"})
    menuItemsList = append(menuItemsList, MenuItem{Url:"http://old.marinhobrandao.com/", Id:"menu-old", Label:"Old site"})
    /*menuItemsList = append(menuItemsList, MenuItem{Url:"/snippets/", Id:"menu-snippets", Label:"Snippets"})*/
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

func BlogPostListHandler(c http.ResponseWriter, req *http.Request) {
    log.Println(req.URL)
    data := "{\"posts\":[]}"

    c.Header().Add("Content-Type", "text/json")

    // Posts from database
    blogPostsList, err := cms.GetRecentBlogPosts(dbDefaultConn.DB(systemConf.DBName))
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

func RequireSuperuser(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
    return func (c http.ResponseWriter, req *http.Request) {
        var cookie *http.Cookie
        var err error

        // Checks secret cookie (temporary until support sessions)
        cookie, err = req.Cookie("mbSuperuser")
        if err != nil || cookie.Value == "LetTarLin" {
            // Return error
            http.Error(c, "Unauthorized", http.StatusUnauthorized)
            return
        }

        // Call the encapsulated function
        handler(c, req)
    }
}

func BlogPostAddHandler(c http.ResponseWriter, req *http.Request) {
    log.Println(req.URL)
    c.Header().Add("Content-Type", "text/json")
    var data string
    var err error
    var blogPost cms.BlogPost
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
                url := cms.Slugify(title)
                if len(postValues["Tags"]) > 0 {
                    tags2 := strings.Split(postValues["Tags"][0], ",")
                    for iTag := range tags2 {
                        tags = append(tags, strings.Trim(tags2[iTag], " "))
                    }
                }

                blogPost = cms.BlogPost{Title:title, Content:content, Published:true, Url:url, Author:DEFAULT_AUTHOR, Tags:tags}
                err = cms.InsertNewBlogPost(dbDefaultConn.DB(systemConf.DBName), &blogPost)
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

func BlogPostDeleteHandler(c http.ResponseWriter, req *http.Request) {
    log.Println(req.URL)
    c.Header().Add("Content-Type", "text/json")
    var data string
    var err error
    var blogPost cms.BlogPost

    // Method not allowed
    if req.Method != "POST" {
        http.Error(c, "Invalid method.", http.StatusMethodNotAllowed)
        return
    }

    // Parse arguments
    args := mux.Vars(req)

    // Load blog post
    err = cms.DeleteBlogPost(dbDefaultConn.DB(systemConf.DBName), args["postId"])

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

// Main routine

func main() {
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

    // Server
    r := mux.NewRouter()
    r.HandleFunc("/", HomeHandler)
    r.HandleFunc("/login/", LoginHandler)
    r.HandleFunc("/api/menu/item/", MenuItemsHandler)
    r.HandleFunc("/api/blog/post/", BlogPostListHandler)
    r.HandleFunc("/api/blog/post/add/", RequireSuperuser(BlogPostAddHandler))
    r.HandleFunc("/api/blog/post/{postId:\\w+}/delete/", RequireSuperuser(BlogPostDeleteHandler))
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(systemConf.StaticRoot))))
    http.Handle("/", r)

    // Start serving!
    log.Fatal(http.ListenAndServe(HTTP_ADDRESS, nil))
}

