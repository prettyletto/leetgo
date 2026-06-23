package generator

func p(name string, kind ValueKind) ParamSpec { return ParamSpec{Name: name, Type: kind} }

var problemSpecs = map[int]ProblemSpec{
	// ============================================================
	// Arrays & Hashing
	// ============================================================
	217: {Slug: "contains-duplicate", Params: []ParamSpec{p("nums", KindIntSlice)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "nums": "[]int{1,2,3,1}"}, Expect: "true"},
			{Input: map[string]string{"_name": "example 2", "nums": "[]int{1,2,3,4}"}, Expect: "false"},
			{Input: map[string]string{"_name": "single element", "nums": "[]int{0}"}, Expect: "false"},
		}},
	242: {Slug: "valid-anagram", Params: []ParamSpec{p("s", KindString), p("t", KindString)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "s": `"anagram"`, "t": `"nagaram"`}, Expect: "true"},
			{Input: map[string]string{"_name": "example 2", "s": `"rat"`, "t": `"car"`}, Expect: "false"},
		}},
	1: {Slug: "two-sum", Params: []ParamSpec{p("nums", KindIntSlice), p("target", KindInt)}, Return: ReturnSpec{KindIntSlice}, Comparison: CmpDeep,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "nums": "[]int{2,7,11,15}", "target": "9"}, Expect: "[]int{0,1}"},
			{Input: map[string]string{"_name": "example 2", "nums": "[]int{3,2,4}", "target": "6"}, Expect: "[]int{1,2}"},
			{Input: map[string]string{"_name": "example 3", "nums": "[]int{3,3}", "target": "6"}, Expect: "[]int{0,1}"},
		}},
	49: {Slug: "group-anagrams", Params: []ParamSpec{p("strs", KindStringSlice)}, Return: ReturnSpec{KindStringSliceSlice}, Comparison: CmpUnordered,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "strs": `[]string{"eat","tea","tan","ate","nat","bat"}`}, Expect: `[][]string{{"bat"},{"nat","tan"},{"ate","eat","tea"}}`},
		}},
	347: {Slug: "top-k-frequent-elements", Params: []ParamSpec{p("nums", KindIntSlice), p("k", KindInt)}, Return: ReturnSpec{KindIntSlice}, Comparison: CmpUnordered,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "nums": "[]int{1,1,1,2,2,3}", "k": "2"}, Expect: "[]int{1,2}"},
			{Input: map[string]string{"_name": "example 2", "nums": "[]int{1}", "k": "1"}, Expect: "[]int{1}"},
		}},
	238: {Slug: "product-of-array-except-self", Params: []ParamSpec{p("nums", KindIntSlice)}, Return: ReturnSpec{KindIntSlice}, Comparison: CmpDeep,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "nums": "[]int{1,2,3,4}"}, Expect: "[]int{24,12,8,6}"},
			{Input: map[string]string{"_name": "example 2", "nums": "[]int{-1,1,0,-3,3}"}, Expect: "[]int{0,0,9,0,0}"},
		}},
	36: {Slug: "valid-sudoku", Params: []ParamSpec{p("board", KindByteSliceSlice)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "valid", "board": "[]byte{[]byte(\"53..7....\"),[]byte(\"6..195...\"),[]byte(\".98....6.\"),[]byte(\"8...6...3\"),[]byte(\"4..8.3..1\"),[]byte(\"7...2...6\"),[]byte(\".6....28.\"),[]byte(\"...419..5\"),[]byte(\"....8..79\")}"}, Expect: "true"},
		}},
	128: {Slug: "longest-consecutive-sequence", Params: []ParamSpec{p("nums", KindIntSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "nums": "[]int{100,4,200,1,3,2}"}, Expect: "4"},
			{Input: map[string]string{"_name": "example 2", "nums": "[]int{0,3,7,2,5,8,4,6,0,1}"}, Expect: "9"},
		}},
	271: {Slug: "encode-and-decode-strings", IsDesign: true, DesignNote: "encode-and-decode-strings requires class with Encode/Decode methods"},
	454: {Slug: "4sum-ii", Params: []ParamSpec{p("nums1", KindIntSlice), p("nums2", KindIntSlice), p("nums3", KindIntSlice), p("nums4", KindIntSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "nums1": "[]int{1,2}", "nums2": "[]int{-2,-1}", "nums3": "[]int{-1,2}", "nums4": "[]int{0,2}"}, Expect: "2"},
			{Input: map[string]string{"_name": "example 2", "nums1": "[]int{0}", "nums2": "[]int{0}", "nums3": "[]int{0}", "nums4": "[]int{0}"}, Expect: "1"},
		}},

	// ============================================================
	// Two Pointers
	// ============================================================
	125: {Slug: "valid-palindrome", Params: []ParamSpec{p("s", KindString)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "s": `"A man, a plan, a canal: Panama"`}, Expect: "true"},
			{Input: map[string]string{"_name": "example 2", "s": `"race a car"`}, Expect: "false"},
		}},
	167: {Slug: "two-sum-ii-input-array-is-sorted", Params: []ParamSpec{p("numbers", KindIntSlice), p("target", KindInt)}, Return: ReturnSpec{KindIntSlice}, Comparison: CmpDeep,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "numbers": "[]int{2,7,11,15}", "target": "9"}, Expect: "[]int{1,2}"},
			{Input: map[string]string{"_name": "example 2", "numbers": "[]int{2,3,4}", "target": "6"}, Expect: "[]int{1,3}"},
		}},
	15: {Slug: "3sum", Params: []ParamSpec{p("nums", KindIntSlice)}, Return: ReturnSpec{KindIntSliceSlice}, Comparison: CmpDeep,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "nums": "[]int{-1,0,1,2,-1,-4}"}, Expect: "[][]int{{-1,-1,2},{-1,0,1}}"},
			{Input: map[string]string{"_name": "example 2", "nums": "[]int{0,1,1}"}, Expect: "[][]int{}"},
		}},
	11: {Slug: "container-with-most-water", Params: []ParamSpec{p("height", KindIntSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "height": "[]int{1,8,6,2,5,4,8,3,7}"}, Expect: "49"},
			{Input: map[string]string{"_name": "example 2", "height": "[]int{1,1}"}, Expect: "1"},
		}},
	42: {Slug: "trapping-rain-water", Params: []ParamSpec{p("height", KindIntSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "height": "[]int{0,1,0,2,1,0,1,3,2,1,2,1}"}, Expect: "6"},
		}},

	// ============================================================
	// Sliding Window
	// ============================================================
	121: {Slug: "best-time-to-buy-and-sell-stock", Params: []ParamSpec{p("prices", KindIntSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "prices": "[]int{7,1,5,3,6,4}"}, Expect: "5"},
			{Input: map[string]string{"_name": "example 2", "prices": "[]int{7,6,4,3,1}"}, Expect: "0"},
		}},
	3: {Slug: "longest-substring-without-repeating-characters", Params: []ParamSpec{p("s", KindString)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "s": `"abcabcbb"`}, Expect: "3"},
			{Input: map[string]string{"_name": "example 2", "s": `"bbbbb"`}, Expect: "1"},
		}},
	424: {Slug: "longest-repeating-character-replacement", Params: []ParamSpec{p("s", KindString), p("k", KindInt)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "s": `"ABAB"`, "k": "2"}, Expect: "4"},
		}},
	567: {Slug: "permutation-in-string", Params: []ParamSpec{p("s1", KindString), p("s2", KindString)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "s1": `"ab"`, "s2": `"eidbaooo"`}, Expect: "true"},
		}},
	76: {Slug: "minimum-window-substring", Params: []ParamSpec{p("s", KindString), p("t", KindString)}, Return: ReturnSpec{KindString}, Comparison: CmpExact,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "s": `"ADOBECODEBANC"`, "t": `"ABC"`}, Expect: `"BANC"`},
		}},
	239: {Slug: "sliding-window-maximum", Params: []ParamSpec{p("nums", KindIntSlice), p("k", KindInt)}, Return: ReturnSpec{KindIntSlice}, Comparison: CmpDeep,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "nums": "[]int{1,3,-1,-3,5,3,6,7}", "k": "3"}, Expect: "[]int{3,3,5,5,6,7}"},
		}},

	// ============================================================
	// Stack
	// ============================================================
	20: {Slug: "valid-parentheses", Params: []ParamSpec{p("s", KindString)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "s": `"()"`}, Expect: "true"},
			{Input: map[string]string{"_name": "example 2", "s": `"()[]{}"`}, Expect: "true"},
			{Input: map[string]string{"_name": "example 3", "s": `"(]"`}, Expect: "false"},
		}},
	155: {Slug: "min-stack", IsDesign: true, DesignNote: "min-stack requires class with push/pop/top/getMin methods"},
	150: {Slug: "evaluate-reverse-polish-notation", Params: []ParamSpec{p("tokens", KindStringSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "tokens": `[]string{"2","1","+","3","*"}`}, Expect: "9"},
		}},
	22: {Slug: "generate-parentheses", Params: []ParamSpec{p("n", KindInt)}, Return: ReturnSpec{KindStringSlice}, Comparison: CmpDeep,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "n": "3"}, Expect: `[]string{"((()))","(()())","(())()","()(())","()()()"}`},
		}},
	739: {Slug: "daily-temperatures", Params: []ParamSpec{p("temperatures", KindIntSlice)}, Return: ReturnSpec{KindIntSlice}, Comparison: CmpDeep,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "temperatures": "[]int{73,74,75,71,69,72,76,73}"}, Expect: "[]int{1,1,4,2,1,1,0,0}"},
		}},
	853: {Slug: "car-fleet", Params: []ParamSpec{p("target", KindInt), p("position", KindIntSlice), p("speed", KindIntSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "target": "12", "position": "[]int{10,8,0,5,3}", "speed": "[]int{2,4,1,1,3}"}, Expect: "3"},
		}},
	84: {Slug: "largest-rectangle-in-histogram", Params: []ParamSpec{p("heights", KindIntSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "heights": "[]int{2,1,5,6,2,3}"}, Expect: "10"},
		}},

	// ============================================================
	// Binary Search
	// ============================================================
	704: {Slug: "binary-search", Params: []ParamSpec{p("nums", KindIntSlice), p("target", KindInt)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "found", "nums": "[]int{-1,0,3,5,9,12}", "target": "9"}, Expect: "4"},
			{Input: map[string]string{"_name": "not found", "nums": "[]int{-1,0,3,5,9,12}", "target": "2"}, Expect: "-1"},
		}},
	74: {Slug: "search-a-2d-matrix", Params: []ParamSpec{p("matrix", KindIntSliceSlice), p("target", KindInt)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "matrix": "[][]int{{1,3,5,7},{10,11,16,20},{23,30,34,60}}", "target": "3"}, Expect: "true"},
		}},
	875: {Slug: "koko-eating-bananas", Params: []ParamSpec{p("piles", KindIntSlice), p("h", KindInt)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "piles": "[]int{3,6,7,11}", "h": "8"}, Expect: "4"},
		}},
	153: {Slug: "find-minimum-in-rotated-sorted-array", Params: []ParamSpec{p("nums", KindIntSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "nums": "[]int{3,4,5,1,2}"}, Expect: "1"},
		}},
	33: {Slug: "search-in-rotated-sorted-array", Params: []ParamSpec{p("nums", KindIntSlice), p("target", KindInt)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "nums": "[]int{4,5,6,7,0,1,2}", "target": "0"}, Expect: "4"},
		}},
	981: {Slug: "time-based-key-value-store", IsDesign: true, DesignNote: "time-based-key-value-store requires class with set/get methods"},
	4: {Slug: "median-of-two-sorted-arrays", Params: []ParamSpec{p("nums1", KindIntSlice), p("nums2", KindIntSlice)}, Return: ReturnSpec{KindFloat64}, Comparison: CmpExact,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "nums1": "[]int{1,3}", "nums2": "[]int{2}"}, Expect: "2.0"},
		}},

	// ============================================================
	// Linked List (needs ListNode)
	// ============================================================
	206: {Slug: "reverse-linked-list", NeedsListNode: true, Params: []ParamSpec{p("head", KindListNode)}, Return: ReturnSpec{KindListNode}, Comparison: CmpDeep,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "head": "&ListNode{1, &ListNode{2, &ListNode{3, &ListNode{4, &ListNode{5, nil}}}}}"}, Expect: "&ListNode{5, &ListNode{4, &ListNode{3, &ListNode{2, &ListNode{1, nil}}}}}"},
		}},
	21: {Slug: "merge-two-sorted-lists", NeedsListNode: true, Params: []ParamSpec{p("list1", KindListNode), p("list2", KindListNode)}, Return: ReturnSpec{KindListNode}, Comparison: CmpDeep,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "list1": "&ListNode{1, &ListNode{2, &ListNode{4, nil}}}", "list2": "&ListNode{1, &ListNode{3, &ListNode{4, nil}}}"}, Expect: "&ListNode{1, &ListNode{1, &ListNode{2, &ListNode{3, &ListNode{4, &ListNode{4, nil}}}}}}"},
		}},
	141: {Slug: "linked-list-cycle", NeedsListNode: true, Params: []ParamSpec{p("head", KindListNode)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "no cycle", "head": "&ListNode{1, &ListNode{2, nil}}"}, Expect: "false"},
		}},
	143: {Slug: "reorder-list", NeedsListNode: true, Params: []ParamSpec{p("head", KindListNode)}, Return: ReturnSpec{KindVoid}, Comparison: CmpDeep,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "head": "&ListNode{1, &ListNode{2, &ListNode{3, &ListNode{4, nil}}}}"}, Expect: "&ListNode{1, &ListNode{4, &ListNode{2, &ListNode{3, nil}}}}"},
		}},
	19: {Slug: "remove-nth-node-from-end-of-list", NeedsListNode: true, Params: []ParamSpec{p("head", KindListNode), p("n", KindInt)}, Return: ReturnSpec{KindListNode}, Comparison: CmpDeep,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "head": "&ListNode{1, &ListNode{2, &ListNode{3, &ListNode{4, &ListNode{5, nil}}}}}", "n": "2"}, Expect: "&ListNode{1, &ListNode{2, &ListNode{3, &ListNode{5, nil}}}}"},
		}},
	138: {Slug: "copy-list-with-random-pointer", IsDesign: true, DesignNote: "requires Node struct with Next and Random pointers"},
	2: {Slug: "add-two-numbers", NeedsListNode: true, Params: []ParamSpec{p("l1", KindListNode), p("l2", KindListNode)}, Return: ReturnSpec{KindListNode}, Comparison: CmpDeep,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "l1": "&ListNode{2, &ListNode{4, &ListNode{3, nil}}}", "l2": "&ListNode{5, &ListNode{6, &ListNode{4, nil}}}"}, Expect: "&ListNode{7, &ListNode{0, &ListNode{8, nil}}}"},
		}},
	287: {Slug: "find-the-duplicate-number", Params: []ParamSpec{p("nums", KindIntSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "nums": "[]int{1,3,4,2,2}"}, Expect: "2"},
		}},
	146: {Slug: "lru-cache", IsDesign: true, DesignNote: "lru-cache requires class with Get/Put methods"},
	23: {Slug: "merge-k-sorted-lists", NeedsListNode: true, Params: []ParamSpec{p("lists", KindListNode)}, Return: ReturnSpec{KindListNode}, Comparison: CmpSkip,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "skip", "lists": "nil"}, Expect: "nil"},
		}},
	25: {Slug: "reverse-nodes-in-k-group", NeedsListNode: true, Params: []ParamSpec{p("head", KindListNode), p("k", KindInt)}, Return: ReturnSpec{KindListNode}, Comparison: CmpDeep,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "head": "&ListNode{1, &ListNode{2, &ListNode{3, &ListNode{4, &ListNode{5, nil}}}}}", "k": "2"}, Expect: "&ListNode{2, &ListNode{1, &ListNode{4, &ListNode{3, &ListNode{5, nil}}}}}"},
		}},

	// ============================================================
	// Trees
	// ============================================================
	94:  {Slug: "binary-tree-inorder-traversal", NeedsTreeNode: true, Params: []ParamSpec{p("root", KindTreeNode)}, Return: ReturnSpec{KindIntSlice}, Comparison: CmpDeep, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "root": "&TreeNode{1, nil, &TreeNode{2, &TreeNode{3, nil, nil}, nil}}"}, Expect: "[]int{1,3,2}"}}},
	104: {Slug: "maximum-depth-of-binary-tree", NeedsTreeNode: true, Params: []ParamSpec{p("root", KindTreeNode)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "root": "&TreeNode{3, &TreeNode{9, nil, nil}, &TreeNode{20, &TreeNode{15, nil, nil}, &TreeNode{7, nil, nil}}}"}, Expect: "3"}}},
	226: {Slug: "invert-binary-tree", NeedsTreeNode: true, Params: []ParamSpec{p("root", KindTreeNode)}, Return: ReturnSpec{KindTreeNode}, Comparison: CmpDeep, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "root": "&TreeNode{4, &TreeNode{2, &TreeNode{1, nil, nil}, &TreeNode{3, nil, nil}}, &TreeNode{7, &TreeNode{6, nil, nil}, &TreeNode{9, nil, nil}}}"}, Expect: "&TreeNode{4, &TreeNode{7, &TreeNode{9, nil, nil}, &TreeNode{6, nil, nil}}, &TreeNode{2, &TreeNode{3, nil, nil}, &TreeNode{1, nil, nil}}}}"}}},
	101: {Slug: "symmetric-tree", NeedsTreeNode: true, Params: []ParamSpec{p("root", KindTreeNode)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "root": "&TreeNode{1, &TreeNode{2, &TreeNode{3, nil, nil}, &TreeNode{4, nil, nil}}, &TreeNode{2, &TreeNode{4, nil, nil}, &TreeNode{3, nil, nil}}}"}, Expect: "true"}}},
	108: {Slug: "convert-sorted-array-to-binary-search-tree", NeedsTreeNode: true, Params: []ParamSpec{p("root", KindTreeNode)}, Return: ReturnSpec{KindTreeNode}, Comparison: CmpSkip, Examples: []ExampleSpec{{Input: map[string]string{"_name": "skip", "root": "nil"}, Expect: "nil"}}},
	100: {Slug: "same-tree", NeedsTreeNode: true, Params: []ParamSpec{p("p", KindTreeNode), p("q", KindTreeNode)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool, Examples: []ExampleSpec{{Input: map[string]string{"_name": "same", "p": "&TreeNode{1, &TreeNode{2, nil, nil}, &TreeNode{3, nil, nil}}", "q": "&TreeNode{1, &TreeNode{2, nil, nil}, &TreeNode{3, nil, nil}}"}, Expect: "true"}}},
	110: {Slug: "balanced-binary-tree", NeedsTreeNode: true, Params: []ParamSpec{p("root", KindTreeNode)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "root": "&TreeNode{3, &TreeNode{9, nil, nil}, &TreeNode{20, &TreeNode{15, nil, nil}, &TreeNode{7, nil, nil}}}"}, Expect: "true"}}},
	543: {Slug: "diameter-of-binary-tree", NeedsTreeNode: true, Params: []ParamSpec{p("root", KindTreeNode)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "root": "&TreeNode{1, &TreeNode{2, &TreeNode{4, nil, nil}, &TreeNode{5, nil, nil}}, &TreeNode{3, nil, nil}}"}, Expect: "3"}}},
	105: {Slug: "construct-binary-tree-from-preorder-and-inorder-traversal", NeedsTreeNode: true, Params: []ParamSpec{p("preorder", KindIntSlice), p("inorder", KindIntSlice)}, Return: ReturnSpec{KindTreeNode}, Comparison: CmpSkip, Examples: []ExampleSpec{{Input: map[string]string{"_name": "skip", "preorder": "[]int{3,9,20,15,7}", "inorder": "[]int{9,3,15,20,7}"}, Expect: "nil"}}},
	102: {Slug: "binary-tree-level-order-traversal", NeedsTreeNode: true, Params: []ParamSpec{p("root", KindTreeNode)}, Return: ReturnSpec{KindIntSliceSlice}, Comparison: CmpDeep, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "root": "&TreeNode{3, &TreeNode{9, nil, nil}, &TreeNode{20, &TreeNode{15, nil, nil}, &TreeNode{7, nil, nil}}}"}, Expect: "[][]int{{3},{9,20},{15,7}}"}}},
	98:  {Slug: "validate-binary-search-tree", NeedsTreeNode: true, Params: []ParamSpec{p("root", KindTreeNode)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool, Examples: []ExampleSpec{{Input: map[string]string{"_name": "valid", "root": "&TreeNode{2, &TreeNode{1, nil, nil}, &TreeNode{3, nil, nil}}"}, Expect: "true"}}},
	230: {Slug: "kth-smallest-element-in-a-bst", NeedsTreeNode: true, Params: []ParamSpec{p("root", KindTreeNode), p("k", KindInt)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "root": "&TreeNode{3, &TreeNode{1, nil, &TreeNode{2, nil, nil}}, &TreeNode{4, nil, nil}}", "k": "1"}, Expect: "1"}}},
	199: {Slug: "binary-tree-right-side-view", NeedsTreeNode: true, Params: []ParamSpec{p("root", KindTreeNode)}, Return: ReturnSpec{KindIntSlice}, Comparison: CmpDeep, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "root": "&TreeNode{1, &TreeNode{2, nil, &TreeNode{5, nil, nil}}, &TreeNode{3, nil, &TreeNode{4, nil, nil}}}"}, Expect: "[]int{1,3,4}"}}},
	124: {Slug: "binary-tree-maximum-path-sum", NeedsTreeNode: true, Params: []ParamSpec{p("root", KindTreeNode)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "root": "&TreeNode{1, &TreeNode{2, nil, nil}, &TreeNode{3, nil, nil}}"}, Expect: "6"}}},
	297: {Slug: "serialize-and-deserialize-binary-tree", IsDesign: true, DesignNote: "requires Codec class with serialize/deserialize methods"},

	// ============================================================
	// Tries (Design problems)
	// ============================================================
	208: {Slug: "implement-trie-prefix-tree", IsDesign: true, DesignNote: "implement-trie-prefix-tree requires Trie class with insert/search/startsWith"},
	211: {Slug: "design-add-and-search-words-data-structure", IsDesign: true, DesignNote: "requires WordDictionary class with addWord/search"},
	212: {Slug: "word-search-ii", Params: []ParamSpec{p("board", KindByteSliceSlice), p("words", KindStringSlice)}, Return: ReturnSpec{KindStringSlice}, Comparison: CmpDeep,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "board": `[][]byte{{'o','a','a','n'},{'e','t','a','e'},{'i','h','k','r'},{'i','f','l','v'}}`, "words": `[]string{"oath","pea","eat","rain"}`}, Expect: `[]string{"eat","oath"}`},
		}},

	// ============================================================
	// Heap / Priority Queue
	// ============================================================
	703:  {Slug: "kth-largest-element-in-a-stream", IsDesign: true, DesignNote: "requires KthLargest class with constructor and add method"},
	1046: {Slug: "last-stone-weight", Params: []ParamSpec{p("stones", KindIntSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "stones": "[]int{2,7,4,1,8,1}"}, Expect: "1"}}},
	973:  {Slug: "k-closest-points-to-origin", Params: []ParamSpec{p("points", KindIntSliceSlice), p("k", KindInt)}, Return: ReturnSpec{KindIntSliceSlice}, Comparison: CmpDeep, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "points": "[][]int{{1,3},{-2,2}}", "k": "1"}, Expect: "[][]int{{-2,2}}"}}},
	215:  {Slug: "kth-largest-element-in-an-array", Params: []ParamSpec{p("nums", KindIntSlice), p("k", KindInt)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "nums": "[]int{3,2,1,5,6,4}", "k": "2"}, Expect: "5"}}},
	621:  {Slug: "task-scheduler", Params: []ParamSpec{p("tasks", KindByteSlice), p("n", KindInt)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "tasks": `[]byte{'A','A','A','B','B','B'}`, "n": "2"}, Expect: "8"}}},
	355:  {Slug: "design-twitter", IsDesign: true, DesignNote: "requires Twitter class with postTweet/getNewsFeed/follow/unfollow"},
	295:  {Slug: "find-median-from-data-stream", IsDesign: true, DesignNote: "requires MedianFinder class with addNum/findMedian"},

	// ============================================================
	// Backtracking
	// ============================================================
	78:  {Slug: "subsets", Params: []ParamSpec{p("nums", KindIntSlice)}, Return: ReturnSpec{KindIntSliceSlice}, Comparison: CmpUnordered, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "nums": "[]int{1,2,3}"}, Expect: "[][]int{{},{1},{2},{1,2},{3},{1,3},{2,3},{1,2,3}}"}}},
	39:  {Slug: "combination-sum", Params: []ParamSpec{p("candidates", KindIntSlice), p("target", KindInt)}, Return: ReturnSpec{KindIntSliceSlice}, Comparison: CmpDeep, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "candidates": "[]int{2,3,6,7}", "target": "7"}, Expect: "[][]int{{2,2,3},{7}}"}}},
	46:  {Slug: "permutations", Params: []ParamSpec{p("nums", KindIntSlice)}, Return: ReturnSpec{KindIntSliceSlice}, Comparison: CmpDeep, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "nums": "[]int{1,2,3}"}, Expect: "[][]int{{1,2,3},{1,3,2},{2,1,3},{2,3,1},{3,1,2},{3,2,1}}"}}},
	90:  {Slug: "subsets-ii", Params: []ParamSpec{p("nums", KindIntSlice)}, Return: ReturnSpec{KindIntSliceSlice}, Comparison: CmpUnordered, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "nums": "[]int{1,2,2}"}, Expect: "[][]int{{},{1},{1,2},{1,2,2},{2},{2,2}}"}}},
	40:  {Slug: "combination-sum-ii", Params: []ParamSpec{p("candidates", KindIntSlice), p("target", KindInt)}, Return: ReturnSpec{KindIntSliceSlice}, Comparison: CmpDeep, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "candidates": "[]int{10,1,2,7,6,1,5}", "target": "8"}, Expect: "[][]int{{1,1,6},{1,2,5},{1,7},{2,6}}"}}},
	17:  {Slug: "letter-combinations-of-a-phone-number", Params: []ParamSpec{p("digits", KindString)}, Return: ReturnSpec{KindStringSlice}, Comparison: CmpDeep, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "digits": `"23"`}, Expect: `[]string{"ad","ae","af","bd","be","bf","cd","ce","cf"}`}}},
	79:  {Slug: "word-search", Params: []ParamSpec{p("board", KindByteSliceSlice), p("word", KindString)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "board": `[][]byte{{'A','B','C','E'},{'S','F','C','S'},{'A','D','E','E'}}`, "word": `"ABCCED"`}, Expect: "true"}}},
	131: {Slug: "palindrome-partitioning", Params: []ParamSpec{p("s", KindString)}, Return: ReturnSpec{KindStringSliceSlice}, Comparison: CmpDeep, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "s": `"aab"`}, Expect: `[][]string{{"a","a","b"},{"aa","b"}}`}}},
	51:  {Slug: "n-queens", Params: []ParamSpec{p("n", KindInt)}, Return: ReturnSpec{KindStringSliceSlice}, Comparison: CmpDeep, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "n": "4"}, Expect: `[][]string{{".Q..","...Q","Q...","..Q."},{"..Q.","Q...","...Q",".Q.."}}`}}},

	// ============================================================
	// Graphs
	// ============================================================
	200: {Slug: "number-of-islands", Params: []ParamSpec{p("grid", KindByteSliceSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "grid": `[][]byte{{'1','1','1','1','0'},{'1','1','0','1','0'},{'1','1','0','0','0'},{'0','0','0','0','0'}}`}, Expect: "1"}}},
	133: {Slug: "clone-graph", IsDesign: true, DesignNote: "clone-graph requires Node struct with Neighbors slice"},
	695: {Slug: "max-area-of-island", Params: []ParamSpec{p("grid", KindIntSliceSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "grid": "[][]int{{0,0,1,0,0,0,0,1,0,0,0,0,0},{0,0,0,0,0,0,0,1,1,1,0,0,0},{0,1,1,0,1,0,0,0,0,0,0,0,0},{0,1,0,0,1,1,0,0,1,0,1,0,0},{0,1,0,0,1,1,0,0,1,1,1,0,0},{0,0,0,0,0,0,0,0,0,0,1,0,0},{0,0,0,0,0,0,0,1,1,1,0,0,0},{0,0,0,0,0,0,0,1,1,0,0,0,0}}"}, Expect: "6"}}},
	417: {Slug: "pacific-atlantic-water-flow", Params: []ParamSpec{p("heights", KindIntSliceSlice)}, Return: ReturnSpec{KindIntSliceSlice}, Comparison: CmpDeep, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "heights": "[][]int{{1,2,2,3,5},{3,2,3,4,4},{2,4,5,3,1},{6,7,1,4,5},{5,1,1,2,4}}"}, Expect: "[][]int{{0,4},{1,3},{1,4},{2,2},{3,0},{3,1},{4,0}}"}}},
	994: {Slug: "rotting-oranges", Params: []ParamSpec{p("grid", KindIntSliceSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "grid": "[][]int{{2,1,1},{1,1,0},{0,1,1}}"}, Expect: "4"}}},
	207: {Slug: "course-schedule", Params: []ParamSpec{p("numCourses", KindInt), p("prerequisites", KindIntSliceSlice)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "numCourses": "2", "prerequisites": "[][]int{{1,0}}"}, Expect: "true"}}},
	210: {Slug: "course-schedule-ii", Params: []ParamSpec{p("numCourses", KindInt), p("prerequisites", KindIntSliceSlice)}, Return: ReturnSpec{KindIntSlice}, Comparison: CmpDeep, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "numCourses": "2", "prerequisites": "[][]int{{1,0}}"}, Expect: "[]int{0,1}"}}},
	684: {Slug: "redundant-connection", Params: []ParamSpec{p("edges", KindIntSliceSlice)}, Return: ReturnSpec{KindIntSlice}, Comparison: CmpDeep, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "edges": "[][]int{{1,2},{1,3},{2,3}}"}, Expect: "[]int{2,3}"}}},
	323: {Slug: "number-of-connected-components-in-an-undirected-graph", Params: []ParamSpec{p("n", KindInt), p("edges", KindIntSliceSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "n": "5", "edges": "[][]int{{0,1},{1,2},{3,4}}"}, Expect: "2"}}},
	261: {Slug: "graph-valid-tree", Params: []ParamSpec{p("n", KindInt), p("edges", KindIntSliceSlice)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool, Examples: []ExampleSpec{{Input: map[string]string{"_name": "valid", "n": "5", "edges": "[][]int{{0,1},{0,2},{0,3},{1,4}}"}, Expect: "true"}}},
	127: {Slug: "word-ladder", Params: []ParamSpec{p("beginWord", KindString), p("endWord", KindString), p("wordList", KindStringSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "beginWord": `"hit"`, "endWord": `"cog"`, "wordList": `[]string{"hot","dot","dog","lot","log","cog"}`}, Expect: "5"}}},

	// ============================================================
	// Advanced Graphs
	// ============================================================
	332: {Slug: "reconstruct-itinerary", Params: []ParamSpec{p("tickets", KindStringSliceSlice)}, Return: ReturnSpec{KindStringSlice}, Comparison: CmpDeep, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "tickets": `[][]string{{"MUC","LHR"},{"JFK","MUC"},{"SFO","SJC"},{"LHR","SFO"}}`}, Expect: `[]string{"JFK","MUC","LHR","SFO","SJC"}`}}},
	743: {Slug: "network-delay-time", Params: []ParamSpec{p("times", KindIntSliceSlice), p("n", KindInt), p("k", KindInt)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "times": "[][]int{{2,1,1},{2,3,1},{3,4,1}}", "n": "4", "k": "2"}, Expect: "2"}}},
	778: {Slug: "swim-in-rising-water", Params: []ParamSpec{p("grid", KindIntSliceSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "grid": "[][]int{{0,2},{1,3}}"}, Expect: "3"}}},
	787: {Slug: "cheapest-flights-within-k-stops", Params: []ParamSpec{p("n", KindInt), p("flights", KindIntSliceSlice), p("src", KindInt), p("dst", KindInt), p("k", KindInt)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "n": "4", "flights": "[][]int{{0,1,100},{1,2,100},{2,0,100},{1,3,600},{2,3,200}}", "src": "0", "dst": "3", "k": "1"}, Expect: "700"}}},

	// ============================================================
	// Dynamic Programming
	// ============================================================
	70:   {Slug: "climbing-stairs", Params: []ParamSpec{p("n", KindInt)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "n": "2"}, Expect: "2"}, {Input: map[string]string{"_name": "example 2", "n": "3"}, Expect: "3"}}},
	198:  {Slug: "house-robber", Params: []ParamSpec{p("nums", KindIntSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "nums": "[]int{1,2,3,1}"}, Expect: "4"}}},
	213:  {Slug: "house-robber-ii", Params: []ParamSpec{p("nums", KindIntSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "nums": "[]int{2,3,2}"}, Expect: "3"}}},
	5:    {Slug: "longest-palindromic-substring", Params: []ParamSpec{p("s", KindString)}, Return: ReturnSpec{KindString}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "s": `"babad"`}, Expect: `"bab"`}}},
	647:  {Slug: "palindromic-substrings", Params: []ParamSpec{p("s", KindString)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "s": `"abc"`}, Expect: "3"}}},
	91:   {Slug: "decode-ways", Params: []ParamSpec{p("s", KindString)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "s": `"12"`}, Expect: "2"}}},
	322:  {Slug: "coin-change", Params: []ParamSpec{p("coins", KindIntSlice), p("amount", KindInt)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "coins": "[]int{1,2,5}", "amount": "11"}, Expect: "3"}}},
	152:  {Slug: "maximum-product-subarray", Params: []ParamSpec{p("nums", KindIntSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "nums": "[]int{2,3,-2,4}"}, Expect: "6"}}},
	139:  {Slug: "word-break", Params: []ParamSpec{p("s", KindString), p("wordDict", KindStringSlice)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "s": `"leetcode"`, "wordDict": `[]string{"leet","code"}`}, Expect: "true"}}},
	300:  {Slug: "longest-increasing-subsequence", Params: []ParamSpec{p("nums", KindIntSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "nums": "[]int{10,9,2,5,3,7,101,18}"}, Expect: "4"}}},
	416:  {Slug: "partition-equal-subset-sum", Params: []ParamSpec{p("nums", KindIntSlice)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "nums": "[]int{1,5,11,5}"}, Expect: "true"}}},
	309:  {Slug: "best-time-to-buy-and-sell-stock-with-cooldown", Params: []ParamSpec{p("prices", KindIntSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "prices": "[]int{1,2,3,0,2}"}, Expect: "3"}}},
	494:  {Slug: "target-sum", Params: []ParamSpec{p("nums", KindIntSlice), p("target", KindInt)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "nums": "[]int{1,1,1,1,1}", "target": "3"}, Expect: "5"}}},
	1143: {Slug: "longest-common-subsequence", Params: []ParamSpec{p("text1", KindString), p("text2", KindString)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "text1": `"abcde"`, "text2": `"ace"`}, Expect: "3"}}},
	72:   {Slug: "edit-distance", Params: []ParamSpec{p("word1", KindString), p("word2", KindString)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "word1": `"horse"`, "word2": `"ros"`}, Expect: "3"}}},
	518:  {Slug: "coin-change-ii", Params: []ParamSpec{p("amount", KindInt), p("coins", KindIntSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "amount": "5", "coins": "[]int{1,2,5}"}, Expect: "4"}}},

	// ============================================================
	// Greedy
	// ============================================================
	53:   {Slug: "maximum-subarray", Params: []ParamSpec{p("nums", KindIntSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "nums": "[]int{-2,1,-3,4,-1,2,1,-5,4}"}, Expect: "6"}}},
	55:   {Slug: "jump-game", Params: []ParamSpec{p("nums", KindIntSlice)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "nums": "[]int{2,3,1,1,4}"}, Expect: "true"}}},
	45:   {Slug: "jump-game-ii", Params: []ParamSpec{p("nums", KindIntSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "nums": "[]int{2,3,1,1,4}"}, Expect: "2"}}},
	134:  {Slug: "gas-station", Params: []ParamSpec{p("gas", KindIntSlice), p("cost", KindIntSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "gas": "[]int{1,2,3,4,5}", "cost": "[]int{3,4,5,1,2}"}, Expect: "3"}}},
	846:  {Slug: "hand-of-straights", Params: []ParamSpec{p("hand", KindIntSlice), p("groupSize", KindInt)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "hand": "[]int{1,2,3,6,2,3,4,7,8}", "groupSize": "3"}, Expect: "true"}}},
	1899: {Slug: "merge-triplets-to-form-target-triplet", Params: []ParamSpec{p("triplets", KindIntSliceSlice), p("target", KindIntSlice)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "triplets": "[][]int{{2,5,3},{1,8,4},{1,7,5}}", "target": "[]int{2,7,5}"}, Expect: "true"}}},
	678:  {Slug: "valid-parenthesis-string", Params: []ParamSpec{p("s", KindString)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "s": `"()"`}, Expect: "true"}}},

	// ============================================================
	// Intervals
	// ============================================================
	56:  {Slug: "merge-intervals", Params: []ParamSpec{p("intervals", KindIntSliceSlice)}, Return: ReturnSpec{KindIntSliceSlice}, Comparison: CmpDeep, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "intervals": "[][]int{{1,3},{2,6},{8,10},{15,18}}"}, Expect: "[][]int{{1,6},{8,10},{15,18}}"}}},
	57:  {Slug: "insert-interval", Params: []ParamSpec{p("intervals", KindIntSliceSlice), p("newInterval", KindIntSlice)}, Return: ReturnSpec{KindIntSliceSlice}, Comparison: CmpDeep, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "intervals": "[][]int{{1,3},{6,9}}", "newInterval": "[]int{2,5}"}, Expect: "[][]int{{1,5},{6,9}}"}}},
	435: {Slug: "non-overlapping-intervals", Params: []ParamSpec{p("intervals", KindIntSliceSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "intervals": "[][]int{{1,2},{2,3},{3,4},{1,3}}"}, Expect: "1"}}},
	252: {Slug: "meeting-rooms", Params: []ParamSpec{p("intervals", KindIntSliceSlice)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "intervals": "[][]int{{0,30},{5,10},{15,20}}"}, Expect: "false"}}},
	253: {Slug: "meeting-rooms-ii", Params: []ParamSpec{p("intervals", KindIntSliceSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "intervals": "[][]int{{0,30},{5,10},{15,20}}"}, Expect: "2"}}},
	986: {Slug: "interval-list-intersections", Params: []ParamSpec{p("firstList", KindIntSliceSlice), p("secondList", KindIntSliceSlice)}, Return: ReturnSpec{KindIntSliceSlice}, Comparison: CmpDeep, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "firstList": "[][]int{{0,2},{5,10},{13,23},{24,25}}", "secondList": "[][]int{{1,5},{8,12},{15,24},{25,26}}"}, Expect: "[][]int{{1,2},{5,5},{8,10},{15,23},{24,24},{25,25}}"}}},

	// ============================================================
	// Math & Geometry
	// ============================================================
	48: {Slug: "rotate-image", Params: []ParamSpec{p("matrix", KindIntSliceSlice)}, Return: ReturnSpec{KindVoid}, Comparison: CmpDeep,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "matrix": "[][]int{{1,2,3},{4,5,6},{7,8,9}}"}, Expect: "[][]int{{7,4,1},{8,5,2},{9,6,3}}"},
		}},
	54: {Slug: "spiral-matrix", Params: []ParamSpec{p("matrix", KindIntSliceSlice)}, Return: ReturnSpec{KindIntSlice}, Comparison: CmpDeep, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "matrix": "[][]int{{1,2,3},{4,5,6},{7,8,9}}"}, Expect: "[]int{1,2,3,6,9,8,7,4,5}"}}},
	73: {Slug: "set-matrix-zeroes", Params: []ParamSpec{p("matrix", KindIntSliceSlice)}, Return: ReturnSpec{KindVoid}, Comparison: CmpDeep,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "matrix": "[][]int{{1,1,1},{1,0,1},{1,1,1}}"}, Expect: "[][]int{{1,0,1},{0,0,0},{1,0,1}}"},
		}},
	202: {Slug: "happy-number", Params: []ParamSpec{p("n", KindInt)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "n": "19"}, Expect: "true"}}},
	66:  {Slug: "plus-one", Params: []ParamSpec{p("digits", KindIntSlice)}, Return: ReturnSpec{KindIntSlice}, Comparison: CmpDeep, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "digits": "[]int{1,2,3}"}, Expect: "[]int{1,2,4}"}}},
	50:  {Slug: "powx-n", Params: []ParamSpec{p("x", KindFloat64), p("n", KindInt)}, Return: ReturnSpec{KindFloat64}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "x": "2.0", "n": "10"}, Expect: "1024.0"}}},
	43:  {Slug: "multiply-strings", Params: []ParamSpec{p("num1", KindString), p("num2", KindString)}, Return: ReturnSpec{KindString}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "num1": `"2"`, "num2": `"3"`}, Expect: `"6"`}}},
	286: {Slug: "walls-and-gates", Params: []ParamSpec{p("rooms", KindIntSliceSlice)}, Return: ReturnSpec{KindVoid}, Comparison: CmpDeep,
		Examples: []ExampleSpec{
			{Input: map[string]string{"_name": "example 1", "rooms": "[][]int{{2147483647,-1,0,2147483647},{2147483647,2147483647,2147483647,-1},{2147483647,-1,2147483647,-1},{0,-1,2147483647,2147483647}}"}, Expect: "[][]int{{3,-1,0,1},{2,2,1,-1},{1,-1,2,-1},{0,-1,3,4}}"},
		}},

	// ============================================================
	// hard-mode only
	// ============================================================
	10:  {Slug: "regular-expression-matching", Params: []ParamSpec{p("s", KindString), p("p", KindString)}, Return: ReturnSpec{KindBool}, Comparison: CmpBool, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "s": `"aa"`, "p": `"a"`}, Expect: "false"}}},
	312: {Slug: "burst-balloons", Params: []ParamSpec{p("nums", KindIntSlice)}, Return: ReturnSpec{KindInt}, Comparison: CmpExact, Examples: []ExampleSpec{{Input: map[string]string{"_name": "example 1", "nums": "[]int{3,1,5,8}"}, Expect: "167"}}},
}
