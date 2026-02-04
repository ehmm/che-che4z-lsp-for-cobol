#!/usr/bin/env python3
import json
import sys
import argparse
import os

def main():
    parser = argparse.ArgumentParser(description='Convert Grapher JSONL output to Graphviz DOT format.')
    parser.add_argument('input', help='Input JSONL file path')
    parser.add_argument('output', help='Output DOT file path')
    args = parser.parse_args()

    if not os.path.exists(args.input):
        print(f"Error: Input file {args.input} not found.")
        sys.exit(1)

    nodes = {}
    edges = []
    program_name = "COBOL_Graph"

    with open(args.input, 'r') as f:
        for line in f:
            if not line.strip():
                continue
            try:
                data = json.loads(line)
                
                if data.get("type") == "program":
                    program_name = data.get("name", program_name)
                
                elif data.get("type") == "node":
                    node_id = data.get("id")
                    # Handle both standard nodes and call nodes
                    name = data.get("name") or data.get("targetName") or f"Node_{node_id}"
                    kind = data.get("kind", "unknown")
                    nodes[node_id] = {"name": name, "kind": kind}
                
                elif data.get("type") == "edge":
                    edges.append((data.get("from"), data.get("to")))
            except json.JSONDecodeError:
                continue

    with open(args.output, 'w') as f:
        f.write(f'digraph "{program_name}"' + " {{\n")
        f.write('  rankdir=TB;')
        f.write('  node [fontname="Courier", fontsize=10, shape=box, style=filled, fillcolor=white];\n')
        f.write('  edge [fontname="Courier", fontsize=8];\n')

        # Write nodes
        for node_id, info in nodes.items():
            label = info["name"]
            kind = info["kind"]
            
            # Styling based on kind
            color = "white"
            shape = "box"
            if kind == "call":
                color = "lightblue"
                shape = "component"
            elif kind == "section":
                color = "lightgrey"
            
            f.write(f'  {node_id} [label="{label}({kind})", fillcolor="{color}", shape="{shape}"];\n')

        # Write edges
        for start, end in edges:
            f.write(f'  {start} -> {end};\n')

        f.write("}" + "\n")

    print(f"Graphviz DOT file generated at: {args.output}")
    print(f"To generate an image, run: dot -Tpng {args.output} -o graph.png")

if __name__ == "__main__":
    main()
