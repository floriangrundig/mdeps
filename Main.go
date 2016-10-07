package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"path/filepath"

	"github.com/deckarep/golang-set"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	// Source directory to walk recursively and search for files
	SourceDir              = kingpin.Arg("source", "Source Directory").Default(".").String()
	SourceFileExtensions   = kingpin.Flag("extensions", "Source File Extension").Short('e').Default(".elm").String()
	ModuleIdRegEx          = kingpin.Flag("moduleIdRegEx", "RegEx with one group for matching module id").Short('i').Default("module ([a-zA-Z0-9\\.]+) ").String()
	DependencyRegEx        = kingpin.Flag("dependencyRegEx", "RegEx with one group for matching dependency id").Default("import ([a-zA-Z0-9\\.]+) ").String()
	IgnoreDependencyRegEx  = kingpin.Flag("ignoreDependencyRegEx", "RegEx which is matched against dependency id and will be ignores").String()
	ReplaceDependencyRegEx = kingpin.Flag("replaceDependencyRegEx", "RegEx which is matched against dependency id and will be replaced").String()
	MaxDepth               = kingpin.Flag("depth", "max level of nesting (TODO better explanation)").Short('n').Int()
	DotOutputFormat        = kingpin.Flag("dot", "dot output format").Short('d').Bool()
	Neo4jOutputFormat      = kingpin.Flag("neo4j", "neo4j output format").Bool()
	DotDiagramTitle        = kingpin.Flag("dotDiagramTitle", "Title of the dot diagram").Default("Dependencies").String()

	regExModuleId        *regexp.Regexp
	regExDependency      *regexp.Regexp
	regExMaxDepth        *regexp.Regexp
	nodeNameReplacements []NodeNameReplacement
	depTree              map[string]*Node
)

type Node struct {
	id           string
	dependencies mapset.Set
}

type NodeNameReplacement struct {
	to   string
	from string
}

func main() {
	kingpin.Parse()
	regExModuleId = regexp.MustCompile(*ModuleIdRegEx)
	regExDependency = regexp.MustCompile(*DependencyRegEx)
	regExMaxDepth = regexp.MustCompile("\\.") // TODO make it a cli flg

	nodeNameReplacements = paresNodeNameReplacements()

	depTree = make(map[string]*Node)

	filepath.Walk(*SourceDir, walkFn)

	if *DotOutputFormat {
		fmt.Printf("digraph \"%s\" {\n", *DotDiagramTitle)
		for nodeId, dependencies := range depTree {
			for _, dep := range dependencies.dependencies.ToSlice() {
				fmt.Printf("\"%s\" -> \"%s\"\n", nodeId, dep)
			}
		}
		fmt.Println("}")

	} else if *Neo4jOutputFormat {
		for nodeId, _ := range depTree {
			fmt.Printf("CREATE (%s:Module  {name: '%s'})\n", nodeId, nodeId)
		}
		fmt.Println("CREATE")

		counter := 0
		for nodeId, dependencies := range depTree {
			counter++
			deps := dependencies.dependencies.ToSlice()
			for idx, dep := range deps {
				if counter < len(depTree) {
					fmt.Printf("\t(%s)-[:DEPENDS ]->(%s),\n", nodeId, dep)
				} else {
					if idx+1 < len(deps) {
						fmt.Printf("\t(%s)-[:DEPENDS ]->(%s),\n", nodeId, dep)
					} else {
						fmt.Printf("\t(%s)-[:DEPENDS ]->(%s)\n", nodeId, dep)
					}
				}

			}
		}

	} else {
		for nodeId, dependencies := range depTree {
			fmt.Println(nodeId)
			for _, dep := range dependencies.dependencies.ToSlice() {
				fmt.Println("---> ", dep)
			}
		}
	}

}

func walkFn(path string, info os.FileInfo, err error) error {
	var node *Node
	if !info.IsDir() {
		if filepath.Ext(path) == *SourceFileExtensions {
			node = parseFile(path)

			// merge if we have already a node with that id (e.g. due to replacement or cutoff the nesting level)
			n, ok := depTree[node.id]
			if ok {
				for _, d := range node.dependencies.ToSlice() {
					n.dependencies.Add(d)
				}
			} else {
				depTree[node.id] = node
			}
		}
	}
	return nil
}

func parseFile(path string) *Node {

	var id string
	dependencies := mapset.NewSet()

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		matches := regExModuleId.FindStringSubmatch(line)
		if len(matches) == 2 {
			id = transformNodeName(matches[1])
		}

		matches = regExDependency.FindStringSubmatch(line)
		if len(matches) == 2 {
			addDependency := true

			if *IgnoreDependencyRegEx != "" {
				matchedIgnore, err := regexp.MatchString(*IgnoreDependencyRegEx, matches[1])
				if err == nil && matchedIgnore {
					addDependency = false
				}
			}

			if addDependency {
				name := transformNodeName(matches[1])
				// fmt.Println("Dependency for", id, name)
				dependencies.Add(name)
			}
		}

	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return &Node{
		id:           id,
		dependencies: dependencies,
	}
}

func transformNodeName(name string) string {
	for _, replacement := range nodeNameReplacements {

		matchedReplacement, err := regexp.MatchString(replacement.from, name)
		if err == nil && matchedReplacement {
			return replacement.to
		}
	}

	if *MaxDepth > 0 {
		matches := regExMaxDepth.FindAllStringIndex(name, -1)

		if len(matches) > (*MaxDepth - 1) {
			return name[0:matches[*MaxDepth-1][0]]
		}
	}

	return name
}

func paresNodeNameReplacements() []NodeNameReplacement {
	nodeNameReplacements := []NodeNameReplacement{}

	if *ReplaceDependencyRegEx != "" {
		parts := strings.Split(*ReplaceDependencyRegEx, "@@@")

		for _, part := range parts {

			replacement := strings.Split(part, "!!!")
			if len(replacement) == 2 {
				nodeNameReplacements = append(nodeNameReplacements, NodeNameReplacement{
					to:   replacement[0],
					from: replacement[1],
				})
			} else {
				fmt.Println("Error parsing replaceDependencyRegEx", part)
			}
		}
	}
	return nodeNameReplacements
}
