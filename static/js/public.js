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
}

function PageCtrl($scope, $routeParams, $http, $location) {
    $scope.params = $routeParams;

    // Sets a different template for Real Life
    if ($scope.params.pageSlug == 'real-life') {
        $scope.extraTemplate = '/templates/real-life.html';
    }

    // Function to load page data
    $scope.loadPage = function() {
        $http.get('/api/page/by-slug/'+$scope.params.pageSlug+'/')
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

    // Real life photos
    $scope.loadRealLifePhotos = function() {
        $http.get('/api/photo/published/')
            .success(function(data){
                $scope.realLifePhotos = data.photos;
            })
            .error(function(data, status, headers, config) {
                console.log('loadRealLifePhotos', data, status, headers, config); // XXX
            });
    }
    $scope.loadRealLifePhotos();
}

