from typing import List
from collections import deque


class TreeNode:

    def __init__(self, val: int, height: int, left=None, right=None) -> None:
        self.val = val
        self.left = left
        self.right = right
        self.height = height
        

def draw(filename: str, root: TreeNode) -> TreeNode:
    colors = {
        "<RT>RT": "#FDF3D0",
        "<MD>MD": "#DCE8FA",
        "<LF>LF": "#F1CFCD",
    }
    
    with open(filename + ".dot", "w") as f:
        f.write('strict digraph {\n')
        f.write('node [shape=record];\n')

        queue = deque([(None, root, "")])
        while queue:
            parent, node, direction = queue.popleft()
            left = "<L>L" if node.left is not None else ""
            right = "<R>R" if node.right is not None else ""
            nest = "<MD>MD"
            if parent is None:
                nest = "<RT>RT"
            elif node.left is None and node.right is None:
                nest = "<LF>LF"
            
            f.write(f'{node.path} [label="{left}|{{<C>{node.val}|{nest}}}|{right}" style=filled fillcolor="{colors[nest]}"];\n')
            
            if parent is not None:
                f.write(f'{parent.path}:{direction} -> {node.path}:C;\n')
            
            if node.left:
                queue.append((node, node.left, "L"))
            if node.right:
                queue.append((node, node.right, "R"))

        f.write('}\n')
