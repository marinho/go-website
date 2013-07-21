package cms

import (
    //"log"
    "time"
    "strings"
    "labix.org/v2/mgo"
    "labix.org/v2/mgo/bson"
)

const BLOG_POST_COLL_NAME = "blog_posts"
type BlogPost struct {
    Id bson.ObjectId `bson:"_id,omitempty"`
    Slug string
    Title string
    Content string // In Textile markup format
    Published bool
    PubDate time.Time //bson.MongoTimestamp
    Author string
    Tags []string
}

const PAGE_COLL_NAME = "pages"
type Page struct {
    Id bson.ObjectId `bson:"_id,omitempty"`
    Slug string
    Title string
    Content string // In Textile markup format
    Published bool
    PubDate time.Time //bson.MongoTimestamp
    Author string
    Tags []string
}

const PHOTO_COLL_NAME = "photos"
type Photo struct {
    Id bson.ObjectId `bson:"_id,omitempty"`
    Filename string // In Textile markup format
    MimeType string
    Published bool
    PubDate time.Time //bson.MongoTimestamp
    Author string
    Tags []string
}

/* GENERAL */

// Returns the string in small letters, replacing spaces for hifens
func Slugify(s string) string {
    // TODO: clean letters to NFKD: http://play.golang.org/p/D7hmrTwi-i
    for strings.Contains(s, "  ") {
        s = strings.Replace(s, "  ", " ", -1)
    }
    s = strings.Replace(s, " ", "-", -1)
    return strings.ToLower(s)
}

/* BLOG POSTS */

// Returns a list of blog post instances
func GetRecentBlogPosts(db *mgo.Database) ([]BlogPost, error) {
    var blogPosts []BlogPost
    var blogPostColl *mgo.Collection

    // Auto Disptach info objects
    blogPostColl = db.C(BLOG_POST_COLL_NAME)
    query := blogPostColl.Find(bson.M{"published":true}).Sort("-pubdate")

    err := query.All(&blogPosts)
    return blogPosts, err
}

// Returns the Id and the error
func InsertNewBlogPost(db *mgo.Database, post *BlogPost) error {
    var blogPostColl *mgo.Collection
    blogPostColl = db.C(BLOG_POST_COLL_NAME)

    // Default empty fields
    post.Id = bson.NewObjectId()
    post.PubDate = time.Now()

    // Insert
    err := blogPostColl.Insert(post)

    return err
}

// Returns the Id and the error
func UpdateBlogPost(db *mgo.Database, post *BlogPost) error {
    var blogPostColl *mgo.Collection
    blogPostColl = db.C(BLOG_POST_COLL_NAME)

    // Insert
    err := blogPostColl.Update(bson.M{"_id":post.Id}, post)

    return err
}

// Loads and return a blog post from database
func GetBlogPost(db *mgo.Database, postId string) (BlogPost,error) {
    var blogPostColl *mgo.Collection
    blogPostColl = db.C(BLOG_POST_COLL_NAME)

    blogPost := BlogPost{}
    err := blogPostColl.Find(bson.M{"_id":bson.ObjectIdHex(postId)}).One(&blogPost)

    return blogPost, err
}

// Loads and return a blog post from database
func DeleteBlogPost(db *mgo.Database, postId string) error {
    var blogPostColl *mgo.Collection
    blogPostColl = db.C(BLOG_POST_COLL_NAME)

    return blogPostColl.Remove(bson.M{"_id":bson.ObjectIdHex(postId)})
}

/* PAGES */

// Returns a list of blog post instances
func ListPages(db *mgo.Database) ([]Page, error) {
    var pages []Page
    var pageColl *mgo.Collection

    // Auto Disptach info objects
    pageColl = db.C(PAGE_COLL_NAME)
    query := pageColl.Find(bson.M{"published":true}).Sort("title")

    err := query.All(&pages)
    return pages, err
}

// Loads and return a page from database, by ID
func GetPage(db *mgo.Database, pageId string) (Page,error) {
    var pageColl *mgo.Collection
    pageColl = db.C(PAGE_COLL_NAME)

    page := Page{}
    err := pageColl.Find(bson.M{"_id":bson.ObjectIdHex(pageId)}).One(&page)

    return page, err
}

// Loads and return a page from database, by slug
func GetPageBySlug(db *mgo.Database, slug string) (Page,error) {
    var pageColl *mgo.Collection
    pageColl = db.C(PAGE_COLL_NAME)

    page := Page{}
    err := pageColl.Find(bson.M{"slug":slug}).One(&page)

    return page, err
}

// Returns true if a page is found
func PageExists(db *mgo.Database, slug string) bool {
    var pageColl *mgo.Collection
    pageColl = db.C(PAGE_COLL_NAME)

    count, err := pageColl.Find(bson.M{"slug":slug}).Count()
    if err == nil && count >= 1 {
        return true
    }

    return false
}

// Returns the Id and the error
func InsertNewPage(db *mgo.Database, page *Page) error {
    var pageColl *mgo.Collection
    pageColl = db.C(PAGE_COLL_NAME)

    // Default empty fields
    page.Id = bson.NewObjectId()
    page.PubDate = time.Now()

    // Insert
    err := pageColl.Insert(page)

    return err
}

// Returns the Id and the error
func UpdatePage(db *mgo.Database, page *Page) error {
    var pageColl *mgo.Collection
    pageColl = db.C(PAGE_COLL_NAME)

    // Insert
    err := pageColl.Update(bson.M{"_id":page.Id}, page)

    return err
}

// Loads and return a page from database
func DeletePage(db *mgo.Database, pageId string) error {
    var pageColl *mgo.Collection
    pageColl = db.C(PAGE_COLL_NAME)

    return pageColl.Remove(bson.M{"_id":bson.ObjectIdHex(pageId)})
}

/* PHOTOS */

// Inserts a new photo
func InsertNewPhoto(db *mgo.Database, photo *Photo) error {
    var photoColl *mgo.Collection
    photoColl = db.C(PHOTO_COLL_NAME)

    // Default empty fields
    photo.Id = bson.NewObjectId()
    photo.PubDate = time.Now()

    // Insert
    err := photoColl.Insert(photo)

    return err
}

// Returns a list of photos instances
func ListPhotos(db *mgo.Database) ([]Photo, error) {
    var photos []Photo
    var photoColl *mgo.Collection

    // Auto Disptach info objects
    photoColl = db.C(PHOTO_COLL_NAME)
    query := photoColl.Find(bson.M{"published":true}).Sort("-pubdate")

    err := query.All(&photos)
    return photos, err
}

