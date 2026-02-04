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
import { CFASTNode } from "./model/cfast";
import { RangeDto } from "./model/external";

export function createRange(items: CFASTNode[]): RangeDto {
  const startLine = (items[0].location?.start?.line ?? 1) - 1;
  const startChar = (items[0].location?.start?.character ?? 1) - 1;

  const endLine = (items[items.length - 1].location?.end?.line ?? 1) - 1;
  const endChar = (items[items.length - 1].location?.end?.character ?? 1) - 1;
  return {
    start: { line: startLine, character: startChar },
    end: { line: endLine, character: endChar },
  };
}

/**
 * Fast 64-bit hash function (FNV-1a variant) for string fingerprints.
 * Returns a 16-character hex string.
 */
export function hash64(str: string): string {
  let h1 = 0x811c9dc5;
  let h2 = 0xcbf29ce4;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    h1 = Math.imul(h1 ^ char, 16777619);
    h2 = Math.imul(h2 ^ char, 16777619);
  }
  return (h1 >>> 0).toString(16).padStart(8, "0") + (h2 >>> 0).toString(16).padStart(8, "0");
}
