#!/usr/bin/env python3
import json
import sys
import os

def get_stable_id(obj):
    if obj.get("type") == "node":
        loc = obj.get("location", {})
        start = loc.get("start", {})
        end = loc.get("end", {})
        # Identity is based on coordinates and name. 
        # Ignoring URI and Kind because they might vary slightly or be null.
        return (
            start.get("line"),
            start.get("character"),
            end.get("line"),
            end.get("character"),
            obj.get("name")
        )
    return None

def load_jsonl(path):
    data = {"program": None, "nodes": {}, "stable_to_id": {}, "id_to_stable": {}, "edges": set()}
    if not os.path.exists(path):
        print(f"File not found: {path}")
        return data
    
    with open(path, "r") as f:
        for line in f:
            line = line.strip()
            if not line: continue
            try:
                obj = json.loads(line)
                t = obj.get("type")
                if t == "program":
                    data["program"] = obj
                elif t == "node":
                    sid = get_stable_id(obj)
                    data["nodes"][sid] = obj
                    data["stable_to_id"][sid] = obj["id"]
                    data["id_to_stable"][obj["id"]] = sid
                elif t == "edge":
                    data["edges"].add((obj["from"], obj["to"]))
            except Exception as e:
                print(f"Error parsing line in {path}: {e}")
    return data

def main():
    if len(sys.argv) < 3:
        print("Usage: compare_jsonl.py <file1.jsonl> <file2.jsonl>")
        sys.exit(1)

    f1_path = sys.argv[1]
    f2_path = sys.argv[2]

    d1 = load_jsonl(f1_path)
    d2 = load_jsonl(f2_path)

    print(f"Comparing {f1_path} vs {f2_path}")
    print("-" * 40)
    print(f"File 1 Nodes: {len(d1['nodes'])} | Edges: {len(d1['edges'])}")
    print(f"File 2 Nodes: {len(d2['nodes'])} | Edges: {len(d2['edges'])}")
    print("-" * 40)

    # Compare nodes by stable identity
    s1_nodes = set(d1['nodes'].keys())
    s2_nodes = set(d2['nodes'].keys())

    common_nodes = s1_nodes & s2_nodes
    only_in_1_nodes = s1_nodes - s2_nodes
    only_in_2_nodes = s2_nodes - s1_nodes

    print(f"Common Nodes: {len(common_nodes)}")
    print(f"Nodes only in File 1 (Go): {len(only_in_1_nodes)}")
    print(f"Nodes only in File 2 (TS): {len(only_in_2_nodes)}")

    # Translate edges to stable identities for comparison
    def translate_edges(data):
        stable_edges = set()
        for f, t in data["edges"]:
            sf = data["id_to_stable"].get(f)
            st = data["id_to_stable"].get(t)
            if sf and st:
                stable_edges.add((sf, st))
        return stable_edges

    se1 = translate_edges(d1)
    se2 = translate_edges(d2)

    common_edges = se1 & se2
    only_in_1_edges = se1 - se2
    only_in_2_edges = se2 - se1

    print(f"Common Edges: {len(common_edges)}")
    print(f"Edges only in File 1 (Go): {len(only_in_1_edges)}")
    print(f"Edges only in File 2 (TS): {len(only_in_2_edges)}")

    if only_in_2_nodes:
        print("\nSample nodes only in TS:")
        for sid in list(only_in_2_nodes)[:5]:
            node = d2['nodes'][sid]
            print(f"  {node.get('name')} at {sid[0]}:{sid[1]}")

    if only_in_2_edges:
        print("\nSample edges only in TS:")
        for sf, st in list(only_in_2_edges)[:10]:
            name_f = d2['nodes'].get(sf, {}).get('name') or d1['nodes'].get(sf, {}).get('name') or "???"
            name_t = d2['nodes'].get(st, {}).get('name') or d1['nodes'].get(st, {}).get('name') or "???"
            print(f"  {name_f} ({sf[0]}:{sf[1]}) -> {name_t} ({st[0]}:{st[1]})")

if __name__ == "__main__":
    main()
