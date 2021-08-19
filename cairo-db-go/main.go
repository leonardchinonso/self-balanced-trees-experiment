package main

import (
	"bufio"
	"fmt"
	"github.com/jmcvetta/randutil"
	"math/rand"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

func getRandomChoices(values []int, weights []int, size int) []int {
	choices := make([]randutil.Choice, len(values), len(values))
	for idx, item := range values {
		choices[idx] = randutil.Choice{Weight: weights[idx], Item: item}
	}

	result := make([]int, size, size)
	for i := 0; i < size; i++ {
		res, err := randutil.WeightedChoice(choices)
		if err != nil {
			panic(err)
		}
		result[i] = res.Item.(int)
	}

	return result
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func convertToHashable(slice []int) string {
	res := ""
	for i, s := range slice {
		res += strconv.Itoa(i) + strconv.Itoa(s)
	}
	return res
}

func sample(upperBound int, sampleSize int) []int {
	sample := make([]int, sampleSize, sampleSize)
	for i := 0; i < sampleSize; i++ {
		sample[i] = rand.Intn(upperBound)
	}
	return sample
}

func sampleFromCountDown(upperBound int, countDown []int) [][]int {
	result := make([][]int, 0, len(countDown))
	for _, k := range countDown {
		sample := sample(upperBound, k)
		result = append(result, sample)
	}
	return result
}

func write(data *string, w *bufio.Writer) {
	_, err := w.WriteString(*data)
	handleError(err)
	err2 := w.Flush()
	handleError(err2)
}

func generateInitialSet() {
	numOfContracts := 1
	numOfAttributes := 1
	numOfFields := 10
	maxNumber := 256    // upper bound on numbers generated as keys and values
	minMappingSize := 1 // minimum number of keys in a mapping
	maxMappingSize := 6 // maximum number of keys in a mapping

	// Assign composite key length to field prefixes
	// 0 means field has a primitive value, 1 - mapping, 2 - mapping of mappings, 3 - mapping of mappings of mapping
	values := []int{0, 1, 2, 3}
	weights := []int{8, 4, 2, 1}
	compositeLengths := getRandomChoices(values, weights, numOfFields)

	// Create a file and defer the file closure
	f, err := os.Create("initial_set.txt")
	handleError(err)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic("Could not close file")
		}
	}(f)

	// Create a buffer writer to batch write to disk
	w := bufio.NewWriter(f)

	for contract := 0; contract < numOfContracts; contract++ {
		for attribute := 0; attribute < numOfAttributes; attribute++ {
			for field := 0; field < numOfFields; field++ {
				compLength := compositeLengths[field]
				if compLength == 0 {
					//Just a primitive value, generate
					v := rand.Intn(maxNumber)
					data := strconv.Itoa(contract) + " " + strconv.Itoa(attribute) + " " + strconv.Itoa(field) + " " + strconv.Itoa(v) + "\n"
					write(&data, w)
				} else {
					// Counting down number of keys that we still need to generate on certain level, 0 means it is done
					countdown := make([]int, compLength, compLength)
					for i := 0; i < compLength; i++ {
						countdown[i] = rand.Intn(maxMappingSize-minMappingSize) + minMappingSize
					}
					// For each part of the composite key, pre-generate keys by sampling without replacement (to avoid duplicates)
					keySamples := sampleFromCountDown(maxNumber, countdown)
					for countdown[0] > 0 {
						// Output field, keys and value
						data := strconv.Itoa(contract) + " " + strconv.Itoa(attribute) + " " + strconv.Itoa(field) + " "
						write(&data, w)

						for i, keySample := range keySamples {
							data := strconv.Itoa(keySample[countdown[i]-1]) + " "
							write(&data, w)
						}

						v := rand.Intn(maxNumber)
						data = strconv.Itoa(v) + "\n"
						write(&data, w)

						// Move on to the next sequence of keys
						i := len(countdown) - 1
						for i >= 0 {
							countdown[i] -= 1
							if countdown[i] > 0 || i == 0 {
								break
							}
							k := rand.Intn(maxMappingSize-minMappingSize) + minMappingSize
							countdown[i] = k
							keySamples[i] = sample(maxNumber, k)
							i -= 1
						}
					}
				}
			}
		}
	}
}

type AVLNode struct {
	key       int       // node key
	composite []int     // composite key of the node
	height    int       // maximum length of paths from the node to any leaves
	nesting   int       // level of sub-tree nesting (0 - contract, 1 - attribute, 2 - field, 3 - key of the field)
	path      string    // path - sequence of 0 (left) or 1 (right) bits specifying how to get from root
	tree      bool      // set to True if this node is root of the tree for the nested data structure
	val       int       // primitive value for the node (mutually exclusive with subtree)
	subtree   []AVLNode // complex value for the node (mutually exclusive with val)
	root      int       // root hash of the subtree that is rooted at this node
}

func convertToIntegers(itemsAsStrings *[][]string) [][]int {
	result := make([][]int, len(*itemsAsStrings))
	for idx, itemLines := range *itemsAsStrings {
		newItemLines := make([]int, len(itemLines), len(itemLines))
		for i := 0; i < len(itemLines); i++ {
			val, err := strconv.Atoi(itemLines[i])
			handleError(err)
			newItemLines[i] = val
		}
		result[idx] = newItemLines
	}
	return result
}

type ByInnerSlice [][]int

func (slice ByInnerSlice) Len() int {
	return len(slice)
}

func (slice ByInnerSlice) Less(i, j int) bool {
	k := 0
	for k < len(slice[i]) || k < len(slice[j]) {
		if k >= len(slice[i]) {
			if k >= len(slice[j]) {
				return false
			} else {
				return true
			}
		}
		if k >= len(slice[j]) {
			return false
		}
		if slice[i][k] < slice[j][k] {
			return true
		} else if slice[i][k] > slice[j][k] {
			return false
		}
		k++
	}
	return false
}

func (slice ByInnerSlice) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// Reads initial set from the file initial_set.txt and build initial Avl tree ensuring it is balanced
func buildInitialTree() []AVLNode {
	itemsAsStrings := readLines("initial_set.txt")
	items := convertToIntegers(&itemsAsStrings)
	sort.Stable(ByInnerSlice(items))

	treeStack := [][]AVLNode{make([]AVLNode, 0)} // Stack of trees being built, put empty tree on the top
	prefixStack := [][]int{make([]int, 0)}       // Stack of the composite key prefixes corresponding to the tree on the tree_stack, put empty prefix on the top

	ans := make([]string, 0)
	for _, item := range items {
		numKeys := len(item) - 1
		// Check if the prefix of the item matches to the prefix on top of the stack
		for numKeys < len(prefixStack[len(prefixStack)-1]) || !reflect.DeepEqual(item[0:len(prefixStack[len(prefixStack)-1])], prefixStack[len(prefixStack)-1]) {
			// Tree on the top of the stack is complete, need to pop it
			prefixStack = prefixStack[:len(prefixStack)-1]
			treeStack = treeStack[:len(treeStack)-1]
			//fmt.Println("prefixStack: ", prefixStack, "treeStack: ", treeStack)
		}
		// Now check if we need to create nested tree
		for numKeys > len(prefixStack[len(prefixStack)-1])+1 {
			nestedTree := make([]AVLNode, 0)
			nestedKey := item[:len(prefixStack[len(prefixStack)-1])+1]
			treeStack[len(treeStack)-1] = append(treeStack[len(treeStack)-1], AVLNode{
				key:       nestedKey[len(prefixStack[len(prefixStack)-1])],
				composite: nestedKey,
				height:    0, // Height is determined during balancing
				nesting:   len(prefixStack) - 1,
				path:      "", // Path is determined during balancing
				tree:      true,
				subtree:   nestedTree,
				val:       0,
				root:      0,
			})
			ans = append(ans, "A")
			treeStack = append(treeStack, nestedTree)
			prefixStack = append(prefixStack, nestedKey)
		}
		// Now simply add a new node to the tree which is on top of the tree stack
		treeStack[len(treeStack)-1] = append(treeStack[len(treeStack)-1], AVLNode{
			key:       item[len(prefixStack[len(prefixStack)-1])],
			composite: item[:len(item)-1],
			height:    0, // Height is determined during balancing
			nesting:   len(prefixStack) - 1,
			path:      "", // Path is determined during balancing
			tree:      false,
			subtree:   nil,
			val:       item[len(item)-1],
			root:      0,
		})
	}
	return treeStack[0]
}

func printTree(nodes *[]AVLNode) {
	for _, node := range *nodes {
		indent := strings.Repeat(" ", node.nesting)
		fmt.Println(node)
		if node.tree {
			fmt.Printf("%s%v %v %v %s", indent, node.nesting, node.key, node.composite, node.path)
		} else {
			fmt.Printf("%s%v %v %v %v %s", indent, node.nesting, node.key, node.composite, node.val, node.path)
		}
	}
}

// splits tree into two subtrees and invokes itself recursively
// for those sub-trees
// the most important "knob" is the choice of the pivot
// the rule we are using for the pivot is as follows
// If binary representation of total number of nodes in the tree starts with `10` (second most significant bit is zero),
// for example, 2, 4, 5, 8, 9, 10, 11, 16, the left subtree is higher than the right subtree, and the size of the left
// subtree is determined by replacing that `10` combination with `01`. For example, for 2 (10), it will be 1 (01),
// for 4 (100) => 2 (010), for 5 (101) => 3 (011), for 8 (1000) => 4 (0100), for 9 (1001) => 5 (0101),
// for 11 (1011) => 7 (0111), for 16 (10000) => 8 (01000). The size of the right subtree can be computed from the
// size of the tree and size of the left subtree.
// If binary representation of total number of nodes in the tree starts `11` (second most significant bit is one),
// for example, 3, 6, 7, 12, 13, 14, 15, the left and the right subtrees are of the equal height. In the case,
// the size of the right subtree is determined by removing the most significant bit from the size of the tree.
// For example, for 3 (11), it will be 1 (1), for 6 (110) => 2 (10), for 7 (111) => 3 (11), for 12 (1100) => 4 (100),
// for 13 (1101) => 5 (101), 14 (1110) => 6 (110), 15 (1111) => 7 (111).
func balanceTree(path string, nodes *[]AVLNode) {
	n := len(*nodes)
	if n == 0 {
		return
	}
	var pivot int
	if n != 1 {
		// we will shift the value of reduced to the right until it becomes either 2 (10) or 3 (11)
		reduced := n
		fullSize := 0
		for reduced > 3 {
			reduced >>= 1
			fullSize = (fullSize << 1) + 1
		}
		if reduced == 3 {
			pivot = (fullSize << 1) + 1
		} else {
			pivot = n - 1 - fullSize
		}
	}
	if (*nodes)[pivot].tree {
		// nested tree
		(*nodes)[pivot].path = path + "M"
		balanceTree(path+"N", &(*nodes)[pivot].subtree)
	} else {
		(*nodes)[pivot].path = path
	}

	left := (*nodes)[:pivot]
	right := (*nodes)[pivot+1:]

	balanceTree(path+"L", &left)
	balanceTree(path+"R", &right)
}

func flattenTree(nodes *[]AVLNode, flat *[]AVLNode) {
	for _, node := range *nodes {
		*flat = append(*flat, node)
		if node.tree {
			flattenTree(&node.subtree, flat)
		}
	}
}

func graphTree(filename string, flat *[]AVLNode) {
	colors := [8]string{"#FDF3D0", "#DCE8FA", "#D9E7D6", "#F1CFCD", "#F5F5F5", "#E1D5E7", "#FFE6CC", "white"}

	f, err := os.Create(filename + ".dot")
	handleError(err)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic("Could not close file")
		}
	}(f)

	w := bufio.NewWriter(f)

	data := "strict digraph {\n"
	write(&data, w)
	data = "node [shape=record];\n"
	write(&data, w)

	paths := make(map[string]bool)
	for _, node := range *flat {
		paths[node.path] = true
	}

	for _, node := range *flat {
		// Find outgoing edges to understand how to display the node
		nextPrefix := node.path
		lenOfNextPrefix := utf8.RuneCountInString(nextPrefix)
		var (
			left  string
			right string
			nest  string
		)

		if strings.HasSuffix(nextPrefix, "M") {
			nextPrefix = nextPrefix[0 : lenOfNextPrefix-1]
			nest = "<N>N"
		} else {
			nest = strconv.Itoa(node.val) // Display node value instead of letter N
		}

		// If there's a left child, whether its a leaf or not, write the <L>L property to the graph
		_, ok1 := paths[nextPrefix+"L"]
		_, ok2 := paths[nextPrefix+"LM"]
		if ok1 || ok2 {
			left = "<L>L"
		}

		// If there's a right child, whether its a leaf or not, write the <R>R property to the graph
		_, ok3 := paths[nextPrefix+"R"]
		_, ok4 := paths[nextPrefix+"RM"]
		if ok3 || ok4 {
			left = "<R>R"
		}

		data = node.path + " " + "[label=\"" + left + "}|{<C>" + strconv.Itoa(node.key) + "|" + nest + "}|" + right + "\"" + " style=filled fillcolor=\"" + colors[node.nesting] + "\"];\n"
		write(&data, w)

		prev := node.path
		lenOfPrev := utf8.RuneCountInString(prev)
		dir := ""
		if strings.HasSuffix(prev, "M") {
			prev = prev[0 : lenOfPrev-1]
		}
		if strings.HasSuffix(prev, "L") || strings.HasSuffix(prev, "R") {
			prevRunes := []rune(prev)
			dir = string(prevRunes[lenOfPrev-1])
			prev = prev[0 : lenOfPrev-1]
			_, ok := paths[prev]
			if prev != "" && !ok {
				prev += "M"
				if _, ok := paths[prev]; !ok {
					panic("path " + prev + " not found")
				}
			} else {
				dir = "N"
				if strings.HasSuffix(prev, "N") {
					prev = prev[0:lenOfPrev-1] + "M"
				}
			}
			if prev != "" {
				data = prev + ":" + dir + " -> " + node.path + ":C;\n"
				write(&data, w)
			}
		}
	}
	data = "}\n"
	write(&data, w)
}

func initialHash(flat *[]AVLNode) int {
	empty, root := hashSubtree()
	if len(empty) != 0 {
		panic("unused tree nodes after computing root hash: " + strconv.Itoa(len(empty)))
	}
	return root
}

func hashSubtree() ([]AVLNode, int) {
	empty := make([]AVLNode, 0)
	return empty, 123456789
}

func readLines(filename string) [][]string {
	f, err := os.Open(filename)
	handleError(err)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic("Could not close file")
		}
	}(f)

	lines := make([][]string, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.Fields(scanner.Text())
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return lines
}

func readFromFile(filename string) []AVLNode {
	items := readLines(filename)
	result := make([]AVLNode, 0)
	for _, itemLine := range items {
		n := len(itemLine)
		path := itemLine[0]
		root, _ := strconv.Atoi(itemLine[1])
		composite := make([]int, 0, n)
		// Add all values except the last to the new composite
		val, _ := strconv.Atoi(itemLine[n-1])
		for i := 2; i < n-1; i++ {
			num, _ := strconv.Atoi(itemLine[i])
			composite = append(composite, num)
		}
		// If not expecting value at the end, add it to the new composite
		if strings.HasSuffix(path, "M") {
			composite = append(composite, val)
			val = 0
		}
		key := composite[len(composite)-1]
		result = append(result, AVLNode{
			key:       key,
			composite: composite,
			height:    0,
			nesting:   len(composite) - 1,
			path:      path,
			tree:      false,
			subtree:   nil,
			val:       val,
			root:      root,
		})
	}

	return result
}

func writeToFile(nodes *[]AVLNode, filename string) {
	f, err := os.Create(filename)
	handleError(err)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic("Could not close file")
		}
	}(f)
	w := bufio.NewWriter(f)

	for _, node := range *nodes {
		nested := strings.HasSuffix(node.path, "M")
		compositeStrings := make([]string, len(node.composite), len(node.composite))
		for idx, val := range node.composite {
			compositeStrings[idx] = strconv.Itoa(val)
		}
		joinedComposites := strings.Join(compositeStrings, " ")
		data := node.path + " 12345 " + joinedComposites + "\n"
		if !nested {
			data = node.path + " 12345 " + joinedComposites + " " + strconv.Itoa(node.val) + "\n"
		}
		write(&data, w)
	}
}

// Randomly selects some nodes from the tree, and also generates some missing nodes
func selectReads(nodes *[]AVLNode, existAmount int, missAmount int) [][]int {
	existIdx := sample(len(*nodes), existAmount)
	result := make([][]int, len(existIdx), len(existIdx))
	for i, idx := range existIdx {
		result[i] = (*nodes)[idx].composite
	}
	dedup := make(map[string]bool)
	for _, r := range result {
		dedup[convertToHashable(r)] = true
	}
	missIdx := sample(len(*nodes), missAmount)

	// some of the missing nodes might accidentally be present, but it is ok
	for _, idx := range missIdx {
		composites := (*nodes)[idx].composite
		key := make([]int, len(composites), len(composites))
		copy(key, composites)
		key[len(key)-1] += 1
		keyStr := convertToHashable(key)
		if _, ok := dedup[keyStr]; !ok {
			dedup[keyStr] = true
			result = append(result, key)
		}
	}

	return result
}

type ByPath []AVLNode

func (nodes ByPath) Len() int {
	return len(nodes)
}

func (nodes ByPath) Less(i, j int) bool {
	return nodes[i].path < nodes[j].path
}

func (nodes ByPath) Swap(i, j int) {
	nodes[i], nodes[j] = nodes[j], nodes[i]
}

func main() {
	//generateInitialSet()
	buildInitialTree()
	//tree := buildInitialTree()

	// Now balance every sub tree to establish correct depth and path values
	//balanceTree("", &tree)
	//
	//flat := make([]AVLNode, 0)
	//flattenTree(&tree, &flat)
	//sort.Stable(ByPath(flat))
	//printTree(&flat)
	//
	//root := initialHash(&flat)
	//fmt.Println("Initial root hash: ", root)
	//
	//sort.Stable(ByPath(flat))
	//writeToFile(&flat, "sorted_hashes.txt")
	//graphTree("initial_graph", &flat)
	// TODO: call subprocess with CMD

	//reads := selectReads(&flat, 4, 1)
	//fmt.Println("selected reads: ", reads)
}
