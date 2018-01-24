package template

// ExcludedDirs is a list of common directorys which are used store application files
// Usually they aren't shipped with the template but could lead to a crash.
var ExcludedDirs = []string{
	"node_modules",
	"bower_components",
	"jspm_packages",
	"dist",
	"build",
	"log",
	"logs",
	"bin",
	"lib",
	"typings",
}
