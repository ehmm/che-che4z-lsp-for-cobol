/*
 * Copyright (c) 2025 Broadcom.
 * The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
 *
 * This program and the accompanying materials are made
 * available under the terms of the Eclipse Public License 2.0
 * which is available at https://www.eclipse.org/legal/epl-2.0/
 *
 * SPDX-License-Identifier: EPL-2.0
 *
 * Contributors:
 *   Broadcom, Inc. - initial API and implementation
 */
import { Location, Paragraph, Program, Section } from "./cfast";
import { Node } from "./Node";
import { GraphDTO } from "./GraphDTO";

export class Graph {
  private programName: string;
  private location: Location;
  private nodes: Map<number, Node> | undefined;
  private edges: Map<number, Set<number>> | undefined;
  private seenNodes = new Set<number>();
  private seenEdges = new Set<string>();

  constructor(
    private head: Program,
    private writer?: (data: any) => void,
  ) {
    this.programName = head.name;
    this.location = head.location;
    if (!this.writer) {
      this.nodes = new Map();
      this.edges = new Map();
    } else {
      this.writer({
        type: "program",
        name: this.programName,
        location: this.location,
      });
    }
  }

  public getProgramName(): string {
    return this.programName;
  }

  public addNode(node: Node): void {
    if (this.writer) {
      if (!this.seenNodes.has(node.id)) {
        this.writer({ ...node, type: "node" });
        this.seenNodes.add(node.id);
      }
      return;
    }
    this.nodes?.set(node.id, node);
  }

  public createNode(node: Paragraph | Section | Program): Node {
    return {
      id: node.id ?? 0,
      parentId: undefined,
      name: (node as Paragraph | Section | Program).name,
      type: node.type,
      details: node.snippet ?? "",
      location: node.location,
    };
  }

  public getNode(id: number | undefined): Node | undefined {
    if (id === undefined) {
      return undefined;
    }
    // In streaming mode, we don't store nodes, but the GraphBuilder 
    // uses this to check if it needs to create a node.
    // We return a dummy node if we've seen the ID to satisfy the builder.
    if (this.writer && this.seenNodes.has(id)) {
      return { id } as Node;
    }
    return this.nodes?.get(id);
  }

  public getAllNodes(): Map<number, Node> {
    return this.nodes ?? new Map();
  }

  public addOrAppendEdge(parentId: number, childId: number) {
    if (this.writer) {
      const key = `${parentId}->${childId}`;
      if (!this.seenEdges.has(key)) {
        this.writer({ type: "edge", from: parentId, to: childId });
        this.seenEdges.add(key);
      }
      return;
    }

    const edge: Set<number> | undefined = this.edges?.get(parentId);
    if (edge && edge.size > 0) {
      edge.add(childId);
      this.edges?.set(parentId, edge);
    } else {
      this.edges?.set(parentId, new Set([childId]));
    }
  }

  public getAllEdges(): Map<number, Set<number>> {
    return this.edges ?? new Map();
  }

  public normalize(): GraphDTO {
    return this.createGraphDTO();
  }

  private createGraphDTO(): GraphDTO {
    return {
      id: this.head.id ?? 0,
      programName: this.programName,
      location: this.location,
      nodes: Array.from(this.nodes?.entries() ?? []),
      edges: this.getEdges(),
    };
  }

  private getEdges(): [number, number[]][] {
    const edges: [number, Set<number>][] = Array.from(
      this.edges?.entries() ?? [],
    );
    const result: [number, number[]][] = [];
    edges.forEach((value: [number, Set<number>]) => {
      result.push([value[0], Array.from(value[1])]);
    });
    return result;
  }

  private programToNode(program: Program): Node {
    return {
      id: 0,
      parentId: undefined,
      name: program.name,
      type: "program",
      details: "",
      location: program.location,
    };
  }
}
