/* Controllers */

function AdminCtrl($scope, $http) {
    $http.get('/api/admin/menu/').success(function(data){
        $scope.menuItems = data.items;
    });
}

function BlogPostCtrl($scope, $http) {
    // Function to update blog post list
    $scope.updateBlogPosts = function() {
        $http.get('/api/blog/post/').success(function(data){
            $scope.blogPosts = data.posts;
        });
    }
    $scope.updateBlogPosts();
       
    // Function to load blog post data
    $scope.getBlogPost = function(postId, callback) {
        $http.get('/api/blog/post/'+postId+'/')
            .success(function(data){
                $scope.blogPost = data.post;
                if (callback) callback();
            });
    }
 
    // New post
    $scope.submitBlogPostForm = function() {
        var params = {
            Title: $scope.blogPost.Title,
            Content: $scope.blogPost.Content,
            Slug: $scope.blogPost.Slug,
            Tags: $scope.blogPost.Tags ? $scope.blogPost.Tags : ""
        };

        var url = $scope.blogPost.Id ? '/api/blog/post/'+$scope.blogPost.Id+'/' : '/api/blog/post/add/';

        $http.post(url, $scope.encodeUrlVars(params)).success(function(data){
            $scope.updateBlogPosts();
            $scope.closeBlogPostForm();
        });
    }

    // Delete a post
    $scope.deleteBlogPost = function(postId) {
        if (confirm("Confirm deletion?")) {
            $http.post('/api/blog/post/'+postId+'/delete/').success(function(data){
                $scope.updateBlogPosts();
            });
        }
    }

    // Modal for form
    $scope.showBlogPostForm = function (postId) {
        if (postId) {
            $scope.getBlogPost(postId, function(){
                $scope.openBlogPostForm = true;
            });
        } else {
            $scope.blogPost = {
                Id: "",
                Title: "",
                Content: "",
                Slug: "",
                Tags: ""
            };
            $scope.openBlogPostForm = true;
        }
        //$scope.openBlogPostForm = true;
    };
    $scope.closeBlogPostForm = function () {
        $scope.openBlogPostForm = false;
    };
}

function PageCtrl($scope, $routeParams, $http, $location) {
    $scope.params = $routeParams;

    // Function to update blog post list
    $scope.updatePages = function() {
        $http.get('/api/page/').success(function(data){
            $scope.pages = data.pages;
        });
    }
    $scope.updatePages();
    
    // Function to load page data
    $scope.getPage = function(pageId, callback) {
        $http.get('/api/page/'+pageId+'/')
            .success(function(data){
                $scope.page = data.page;
                if (callback) callback();
            });
    }
    
    // New page
    $scope.submitPageForm = function() {
        var params = {
            Title: $scope.page.Title,
            Content: $scope.page.Content,
            Slug: $scope.page.Slug,
            Tags: $scope.page.Tags ? $scope.page.Tags : ""
        };

        var url = $scope.page.Id ? '/api/page/'+$scope.page.Id+'/' : '/api/page/add/';

        $http.post(url, $scope.encodeUrlVars(params)).success(function(data){
            $scope.updatePages();
            $scope.closePageForm();
        });
    }

    // Delete a page
    $scope.deletePage = function(pageId) {
        if (confirm("Confirm deletion?")) {
            $http.post('/api/page/'+pageId+'/delete/').success(function(data){
                $scope.updatePages();
            });
        }
    }

    // Modal for form
    $scope.showPageForm = function (pageId) {
        if (pageId) {
            $scope.getPage(pageId, function(){
                $scope.openPageForm = true;
            });
        } else {
            $scope.page = {
                Id: "",
                Title: "",
                Content: "",
                Slug: "",
                Tags: ""
            };
            $scope.openPageForm = true;
        }
    };
    $scope.closePageForm = function () {
        $scope.openPageForm = false;
    };
}

