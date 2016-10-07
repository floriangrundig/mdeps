# MDeps

MDeps is a simple command line tool which inspects your source code and returns the dependencies between your modules/classes etc.

MDeps is currently used to for ELM projects but should work for Java etc as well (as long as one dependency can be identified by one regular expression).

# Why
We're using MDeps for creating a diagram to show the main components of our project/source code and 
how they interact with each other. This is important to identify archtetural design flaws.

# How it works 

Basically you have to define five settings:

#### Depth
Usually your modules/packages might be nested e.g. 

Pages.SearchPage.Foo
Pages.SearchPage.Bar.Baz

Use flag `-n 2` to limit the depth of your components to 2. This will collapse the dependencies to 
Pages.SearchPage 

#### Replace Dependencies Regular expression
Especially for 3rd party libs your not really interested in detail which module/package is used.
You only want to show that there is a dependency to the lib at a whole. 
Therefor you can specify a regex which will replace the module/package name with something shorter:

The regex has the following format:
```
--replaceDependencyRegEx="replacement_1!!!regex_1@@@replacement_2!!!regex_2"
```
You can define multiple replacements separated by "@@@". 

Each replacement you have to define what a match is replaced with and a regex to find the match (separated by `!!!`).

E.g. 

 --replaceDependencyRegEx="Html!!!^(Html)@@@Json!!!^(Json)"

In this regex we replace all Html.Fooo and Html.Bar with Html and all Json.xxx with Json.

The replacement always wins over the Depth param (see above).

#### Ignore Dependencies
Usually you're not interested in dependencies to core libs of you language.
Use --ignoreDependencyRegEx="^(Dict)|^(String)|^(Time)|^(Regex)|^(Maybe)|^(Array)"

#### File Type
Use -e ".elm" to specify the file filter for your source files.

#### Output format

If not specified you'll get something like
```
module1
---> module1_dependency1
---> module1_dependency2
module2
---> module2_dependency1
---> module2_dependency2
```

If -d  specified you'll get a .dot format (GraphViz) which can be used to create a svg-image.


Use the Makefile example to grep your dependencies: 
e.g. 
```
make source=/Users/flg/code/humio/ui/src/elm | grep -E "\"Pages.* -> \"_Json"
```
Will find all modules which are using some Json Libraries... 


# Download/Install
This tool is written in go so download and install it via `go install`.