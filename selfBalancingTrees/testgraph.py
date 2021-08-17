from drawGraph import draw
# from AVLScheme import TreeNode
import subprocess
from typing import List
import math


class TreeNode:

    def __init__(self, val: int, height: int, left=None, right=None) -> None:
        self.val = val
        self.left = left
        self.right = right
        self.height = height


def populate_paths(root: TreeNode, path="N") -> None:
    if root is None:
        return

    root.path = path
    populate_paths(root.left, path + "L")
    populate_paths(root.right, path + "R")


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


rt = build_tree([6, 5, 7, 2, 3, 4, 1, 5, 3, 1, 12, 34, 67])
populate_paths(rt)
print(rt.val, rt.height)

draw("graph", rt)

subprocess.call(['dot', '-Tpng', 'graph.dot', '-o', 'graph.png'])
