/* Application */

var app = angular.module('mbAdmin', ['ui.bootstrap'], function($routeProvider, $locationProvider) {
    $locationProvider.html5Mode(true);

    $routeProvider
        .when('/', {
            templateUrl: '/templates/admin/home.html',
            controller: LoginCtrl
            })
        .when('/404', {
            templateUrl: '/templates/admin/404.html'
            })
        .when('/pages/', {
            templateUrl: '/templates/admin/pages.html',
            controller: PageCtrl
            })
        .when('/blog-posts/', {
            templateUrl: '/templates/admin/blog-posts.html',
            controller: BlogPostCtrl
            })
        .when('/photos/', {
            templateUrl: '/templates/admin/photos.html',
            controller: PhotoCtrl
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
});

