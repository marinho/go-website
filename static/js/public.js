function MenuCtrl($scope, $http) {
    $http.get('/api/menu/item/').success(function(data){
        $scope.menuItems = data.items;
    });
}

function BlogPostCtrl($scope, $http) {
    // Function to check if current session is authenticated with superuser
    $scope.checkIsSuperuser = function() {
        $http.get('/api/is-superuser/').success(function(data){
            $scope.isSuperuser = data == "yes";
        });
    }
    $scope.checkIsSuperuser();

    // Function to update blog post list
    $scope.updateBlogPosts = function() {
        $http.get('/api/blog/post/').success(function(data){
            $scope.blogPosts = data.posts;
        });
    }
    $scope.updateBlogPosts();
    
    // New post
    $scope.submitNewBlogPost = function() {
        var params = {
            Title: this.Title,
            Content: this.Content,
            Tags: this.Tags ? this.Tags : ""
        };

        $http.post('/api/blog/post/add/', encodeUrlVars(params)).success(function(data){
            $scope.updateBlogPosts();
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
}

function encodeUrlVars(obj) {
    // http://stackoverflow.com/a/6566471/1438115
    var str = "";
    for (var key in obj) {
        if (str != "") {
            str += "&";
        }
        str += key + "=" + obj[key];
    }
    return str;
}
