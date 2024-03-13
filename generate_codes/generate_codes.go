package generate_codes

import (
	"container/heap"
)

// example usage and explanation
/*func main() {
	// The task is to implement an algorithm that:
	// 1 - takes a string array as an input
	// 2 - initializes an empty tree per command
	// 3 - builds a tree starting from the bottom,
	//taking the two lowest frequencies as leaves for the subtree,
	//and then assigning this tree to the array/list
	//(so the root node of a subtree is treated as an element
	//with frequency being a sum of leaves frequencies).
	//The process continues until one tree is built.
	// 4 - then binary codes for each string from the input array are generated
	//by counting the levels of the tree and assigning 0 for left and 1 for right node traversal.
	//E.g. 2 levels until the leaf node, and both to the left node generate the code 00.
	//
	// So, this seems like a tree where values are only stored inside leaves,
	//and the least frequent values are stored at the lowest level of the tree (more nodes from the root).
	// This kind of tree can be interpreted as a Huffman tree (Huffman coding problem),
	//but it is not creating just code from a string producing code for the whole string
	//(then it works on every character to compress string).
	//It operates on symbols being strings inside the string array at the input.
	// So implementations should follow this steps:
	//  1 - take a string array as an input
	//  2 - initialize an empty tree per command and 3 - builds a tree starting from the bottom:
	//  This will be done like in Huffman tree - so frequencies will be calculated and stored
	//  in something like hash table and then priority queue (heap) utilized
	//  to take two elements with lowest frequencies in every step.
	//  Pushing to the heap is O(logn), so it should be better than
	//  sorting (at least O(nlogn)) in each iteration of the tree-building loop.
	//  Then as described in step 3.
	//  4 - traverse to each leaf and generate codes.
	//

	commands := []string{"LEFT", "GRAB", "LEFT", "BACK", "LEFT", "BACK", "LEFT"}

	// Get Huffman codes for the commands
	commandCodes := GetCodesFromListOfCommands(commands)

	// Print the generated Huffman codes
	for _, code := range commandCodes {
		fmt.Printf("Command: %-6s | Huffman Code: %s\n", code.Command, code.Code)
	}
}
*/

// Represents a node in the Huffman tree
type Node struct {
	Value       string
	Frequency   int
	Left, Right *Node
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// PriorityQueue implements heap.Interface and holds Nodes.
type PriorityQueue []*Node

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	//Pop should return the lowest priority/frequency (minimum heap), so use "less than" here.
	return pq[i].Frequency < pq[j].Frequency
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Node)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak / gc
	item.index = -1 // for safety / to mark is not in queue
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *Node, value string, priority int) {
	item.Value = value
	item.Frequency = priority
	heap.Fix(pq, item.index)
}

// InitializeHeap initializes the priority queue (heap) properties
func InitializeHeap(frequencyMap map[string]int) *PriorityQueue {
	pq := make(PriorityQueue, len(frequencyMap))
	i := 0
	for str, freq := range frequencyMap {
		pq[i] = &Node{
			Value:     str,
			Frequency: freq,
			index:     i,
		}
		i++
	}

	heap.Init(&pq)
	return &pq
}

// BuildHuffmanTree builds the Huffman tree
func BuildHuffmanTree(pq *PriorityQueue) *Node {
	for pq.Len() > 1 { // when len is 1 -> this is the root node containing all subtrees => whole built tree
		// Pop two nodes with the lowest frequencies
		node1 := heap.Pop(pq).(*Node)
		node2 := heap.Pop(pq).(*Node)

		// Create a new node without value, with combined frequency and set its children
		// left node (lowest freq) and right node (second lowest freq)
		newNode := &Node{
			Frequency: node1.Frequency + node2.Frequency,
			Left:      node1,
			Right:     node2,
		}

		// Push the new node back to the heap so it will be treated as a regular node
		//- this process builds this tree
		heap.Push(pq, newNode)
	}
	// The root of the Huffman tree is now the only element in the priority queue
	root := (*pq)[0]
	return root
}

// generates Huffman codes for each string based on the Huffman tree
// returns map/hash table with {key="command", value="code"}
// this is recursive depth-first traversal/search (DFS)
func generateHuffmanCodesRecursive(root *Node, code string, codes map[string]string) {
	if root == nil {
		return
	}

	// Base case - Leaf node - end of traverse - save code to the codes map
	if root.Left == nil && root.Right == nil {
		codes[root.Value] = code
		return
	}

	// Traverse left and right with updated code
	generateHuffmanCodesRecursive(root.Left, code+"0", codes)
	generateHuffmanCodesRecursive(root.Right, code+"1", codes)
}

// iterative traversal to avoid stack overflow - max number of input commands is not specified
// returns map/hash table with {key="command", value="code"}
func generateHuffmanCodesIterative(root *Node) map[string]string {
	codes := make(map[string]string)
	stack := []*Node{root}
	codeStack := []string{""}

	for len(stack) > 0 {
		node, code := stack[len(stack)-1], codeStack[len(codeStack)-1]
		stack, codeStack = stack[:len(stack)-1], codeStack[:len(codeStack)-1]

		if node == nil {
			continue
		}

		if node.Left == nil && node.Right == nil {
			codes[node.Value] = code
			continue
		}

		stack = append(stack, node.Right, node.Left)
		codeStack = append(codeStack, code+"1", code+"0")
	}

	return codes
}

// GetCodesFromListOfCommands generates Huffman codes for a given list of commands
// Returns a slice of CommandCode, where each element contains a command and its code
func GetCodesFromListOfCommands(commands []string) map[string]string {
	// 1 get commands
	if commands == nil {
		return nil
	}

	// 2 - initializes an empty tree per command
	//This will be done like in Huffman tree - so frequencies will be calculated
	//and then priority queue (min heap) utilized to take two elements with lowest frequencies
	//in every step.

	// Count the frequency of each string in the commands.
	frequencyMap := make(map[string]int)
	for _, cmd := range commands {
		frequencyMap[cmd]++
	}

	// Initialize heap / priority queue
	pq := InitializeHeap(frequencyMap)

	// Print frequency map (optional)
	// fmt.Println("String | Frequency")
	// fmt.Println("-------------------len", len(frequencyMap))
	// for key, value := range frequencyMap {
	// 	fmt.Printf("%-6s | %d\n", key, value)
	// }

	// 3. Build Huffman tree
	root := BuildHuffmanTree(pq)

	// 4. Generate Huffman codes
	//codes := generateHuffmanCodesRecursive(root)
	codes := generateHuffmanCodesIterative(root)

	return codes
}
