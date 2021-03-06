# TODO: Add documentation to code lines using comments
# TODO: Implement join_left function using the join_right function symmetrically
import math
from typing import List
from drawGraph import draw
import subprocess


class TreeNode:

    def __init__(self, val: int, height: int, left=None, right=None) -> None:
        self.val = val
        self.left = left
        self.right = right
        self.height = height


# Returns true if a node has children and false otherwise
def is_leaf(node: TreeNode) -> bool:
    return node and node.left is None and node.right is None


# This function returns the value, left and right children of a node
def expose(node: TreeNode) -> (TreeNode, int, TreeNode):
    if node is None:
        return None, -1, None
    return node.left, node.val, node.right


def rotate_left(node: TreeNode) -> TreeNode:
    _, _, right = expose(node)
    node.right = None
    node.left = None
    right.left = node
    right.height -= 2
    return right


def rotate_right(node: TreeNode) -> TreeNode:
    left, _, _ = expose(node)
    node.right = None
    node.left = None
    left.right = node
    left.height -= 2
    return left


def join(left: TreeNode, val: int, right: TreeNode) -> TreeNode:
    if left.height > right.height + 1:
        return join_right(left, val, right)

    if right.height > left.height + 1:
        return join_left(left, val, right)

    return TreeNode(val, left.height + 1, left, right)


def join_right(left: TreeNode, val: int, right: TreeNode) -> TreeNode:
    left_child_of_left_tree, val_of_left, right_child_of_left_tree = expose(left)
    if right_child_of_left_tree.height <= right.height + 1:
        new_tree = TreeNode(val, right_child_of_left_tree.height + 1, right_child_of_left_tree, right)
        if new_tree.height <= left_child_of_left_tree.height + 1:
            return TreeNode(val_of_left, left_child_of_left_tree.height + 1, left_child_of_left_tree, new_tree)
        return rotate_left(
            TreeNode(val_of_left, left_child_of_left_tree.height + 1, left_child_of_left_tree, rotate_right(new_tree)))

    new_tree = join_right(right_child_of_left_tree, val, right)
    new_tree_prime = TreeNode(val_of_left, left_child_of_left_tree.height + 1, left_child_of_left_tree, new_tree)
    if new_tree.height <= left_child_of_left_tree.height + 1:
        return new_tree_prime
    return rotate_left(new_tree_prime)


def join_left(left: TreeNode, val: int, right: TreeNode) -> TreeNode:
    left_child_of_right_tree, val_of_right, right_child_of_right_tree = expose(right)
    if left_child_of_right_tree.height <= left.height + 1:
        new_tree = TreeNode(val, left.height + 1, left, left_child_of_right_tree)
        if new_tree.height <= right_child_of_right_tree.height + 1:
            return TreeNode(val_of_righ, new_tree.height + 1, new_tree, right_child_of_right_tree)
        return rotate_right(
            TreeNode(val_of_right, new_tree.height + 1, rotate_left(new_tree), right_child_of_right_tree))

    new_tree = join_left(left, val, left_child_of_right_tree)
    new_tree_prime = TreeNode(val, new_tree.height + 1, new_tree, right_child_of_right_tree)
    if new_tree.height <= right_child_of_right_tree.height + 1:
        return new_tree_prime
    return rotate_right(new_tree_prime)


# This function splits a BST in two by the value of k
# Returns a tuple (L, b, R) where L and R are null or valid nodes
# and b is a boolean denoting if k is a value in the BST
def split(root: TreeNode, k: int) -> (TreeNode, bool, TreeNode):
    # If the root node is the only node in the tree, don't proceed
    if is_leaf(root):
        return root, False, root

    root_left_child, root_val, root_right_child = expose(root)

    # If k is the root's value, split the tree by the root and return
    if k == root_val:
        return root_left_child, True, root_right_child

    # If k is less than the root's value, k must be on the left of the root
    # Recursively split the left subtree and return the tuple if a split is found
    if k < root_val:
        left_child_of_left_subtree, k_in_tree, right_child_of_left_subtree = split(root_left_child, k)
        new_right_child = join(right_child_of_left_subtree, root_val, root_right_child)
        return root_left_child, k_in_tree, new_right_child

    # If execution gets here, k is greater than the root's value and k must be on the right of the root
    # Recursively split the right subtree and return the tuple if a split is found
    left_child_of_right_subtree, k_in_tree, right_child_of_right_subtree = split(root_right_child, k)
    new_left_child = join(root_left_child, root_val, left_child_of_right_subtree)
    return new_left_child, k_in_tree, right_child_of_right_subtree


def split_last(root: TreeNode) -> (TreeNode, int):
    left_child, root_val, right_child = expose(root)
    if is_leaf(right_child):
        return left_child, root_val
    new_root, new_value = split_last(right_child)
    return join(left_child, root_val, new_root), new_value


def join2(left_root: TreeNode, right_root: TreeNode) -> TreeNode:
    if is_leaf(left_root):
        return right_root

    new_left_root, new_value = split_last(left_root)
    return join(new_left_root, new_value, right_root)


def insert(root: TreeNode, k: int) -> None:
    left_child, _, right_child = split(root, k)
    join(left_child, k, right_child)


def delete(root: TreeNode, k: int) -> None:
    left_child, _, right_child = split(root, k)
    join2(left_child, right_child)


# BULK OPERATIONS

def union(root1: TreeNode, root2: TreeNode) -> TreeNode:
    if is_leaf(root1):
        return root2
    if is_leaf(root2):
        return root1

    left_child2, root2_val, right_child2 = expose(root2)
    left_child1, root2_val_in_root1_tree, right_child1 = split(root1, root2_val)
    left_tree = union(left_child1, left_child2)
    right_tree = union(right_child1, right_child2)

    return join(left_tree, root2_val, right_tree)


def intersect(root1: TreeNode, root2: TreeNode) -> TreeNode:
    if is_leaf(root1):
        return root1
    if is_leaf(root2):
        return root2

    left_child2, root2_val, right_child2 = expose(root2)
    left_child1, root2_val_in_root1_tree, right_child1 = split(root1, root2_val)
    left_tree = intersect(left_child1, left_child2)
    right_tree = intersect(right_child1, right_child2)

    if root2_val_in_root1_tree:
        return join(left_tree, root2_val, right_tree)

    return join2(left_tree, right_tree)


def difference(root1: TreeNode, root2: TreeNode) -> TreeNode:
    if is_leaf(root1):
        return root1
    if is_leaf(root2):
        return root1

    left_child2, root2_val, right_child2 = expose(root2)
    left_child1, root2_val_in_root1_tree, right_child1 = split(root1, root2_val)
    left_tree = difference(left_child1, left_child2)
    right_tree = difference(right_child1, right_child2)
    return join2(left_tree, right_tree)


def build_tree(arr: List[int]) -> TreeNode:
    arr.sort()
    n = len(arr)
    height = math.floor(math.log(n, 2))
    return _build_tree(arr, height)


def _build_tree(arr: List[int], height: int) -> TreeNode:
    if len(arr) == 0:
        return

    mid = len(arr) // 2

    root = TreeNode(arr[mid], height)
    root.left = _build_tree(arr[:mid], height - 1)
    root.right = _build_tree(arr[mid + 1:], height - 1)

    return root


def populate_paths(root: TreeNode, path="N") -> None:
    if root is None:
        return
    
    root.path = path
    populate_paths(root.left, path + "L")
    populate_paths(root.right, path + "R")


print("tree1:")
tree1 = build_tree([1, 4, 5])
populate_paths(tree1)
draw("tree1", tree1)
subprocess.call(['dot', '-Tpng', 'tree1.dot', '-o', 'tree1.png'])
# print(tree1.val, tree1.path)
# print(tree1.left.val, tree1.left.path)
# print(tree1.right.val, tree1.right.path)
print()
# print(tree1.right.val, tree1.right.path)
# print(tree1.left.left.val, tree1.left.left.path)
# print(tree1.left.right.val, tree1.left.right.path)
# print(tree1.right.left.val, tree1.right.left.path)
# print(tree1.right.right.val, tree1.right.right.path)

print("tree2:")
tree2 = build_tree([2, 3, 7])
populate_paths(tree2)
draw("tree2", tree2)
subprocess.call(['dot', '-Tpng', 'tree2.dot', '-o', 'tree2.png'])
# print(tree2.val, tree2.path)
# print(tree2.left.val, tree2.left.path)
# print(tree2.right.val, tree2.right.path)
# print()


# JOIN
# joined = join(tree1, 4, tree2)
# populate_paths(joined)


# JOIN2
# joined = join2(tree1, tree2)
# populate_paths(joined)


# SPLIT
# node1, b, node2 = split(tree2, 1)
# print(node1.val, b, node2.val)
# print(node1.left.val, b, node2.val)


# UNION
joined = union(tree1, tree2)
populate_paths(joined)
# print(joined.val)
# print(joined.left.val)
# print(joined.right.val)
# print(joined.left.left.val)
# print(joined.left.right.val)
# print(joined.right.left.val)
# print(joined.right.right.val)


# draw("union", joined)
# subprocess.call(['dot', '-Tpng', 'union.dot', '-o', 'union.png'])
