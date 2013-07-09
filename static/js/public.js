/* Application */

var app = angular.module('mbApp', [], function($routeProvider, $locationProvider) {
    $locationProvider.html5Mode(true);

    $routeProvider
        .when('/', {
            templateUrl: '/templates/home.html',
            controller: BlogPostCtrl
        })
        .when('/404', {
            templateUrl: '/templates/404.html'
        })
        .when('/:pageSlug', {
            templateUrl: '/templates/page.html',
            controller: PageCtrl
        })
        .otherwise({redirectTo: '/404'});
});

// Global functions
app.run(function($rootScope) {
    $rootScope.encodeUrlVars = function(obj) {
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

    // Markdown processor
    var converter = new Showdown.converter();
    $rootScope.processMarkdown = function(raw) {
        return converter.makeHtml(raw);
    }
});

/* Controllers */

function MenuCtrl($scope, $http) {
    $http.get('/api/menu/item/').success(function(data){
        $scope.menuItems = data.items;
    });
}

function GeneralCtrl($scope, $http) {
    // Function to check if current session is authenticated with superuser
    $scope.checkIsSuperuser = function() {
        $http.get('/api/is-superuser/').success(function(data){
            $scope.isSuperuser = data == "yes";
        });
    }
    $scope.checkIsSuperuser();
}

function BlogPostCtrl($scope, $http) {
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

        $http.post('/api/blog/post/add/', $scope.encodeUrlVars(params)).success(function(data){
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

function PageCtrl($scope, $routeParams, $http, $location) {
    $scope.params = $routeParams;

    // Function to load page data
    $scope.loadPage = function() {
        $http.get('/api/page/'+$scope.params.pageSlug+'/')
            .success(function(data){
                $scope.pageInfo = data.page;
            })
            .error(function(data, status, headers, config) {
                if (status == 404) {
                    $location.path('/404?url=/'+$scope.params.pageSlug);
                }
            });
    }
    $scope.loadPage();
}

