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
import { GotoInstruction, PerformInstruction, VNCell } from "./instructions";
import { VirtualMachine } from "./vm";

/**
 * Optimizer stops COBOL Virtual Machine if it process the same COBOL instruction with the same VM state
 */
export class IbmOptimizer {
  private stateMap: Map<number, Set<string>>;
  public totalStates: number = 0;

  public constructor() {
    this.stateMap = new Map<number, Set<string>>();
  }

  /**
   * Apply optimizer to checks if VM must be stopped or not
   * @param vm COBOL Virtual Machine
   * @returns true if VM must be stopped and false otherwise
   */
  public apply(vm: VirtualMachine): boolean {
    const currentInstruction = vm.currentInstruction();

    if (
      !currentInstruction ||
      currentInstruction instanceof VNCell ||
      currentInstruction instanceof GotoInstruction ||
      currentInstruction instanceof PerformInstruction
    ) {
      return false;
    }

    const vmState = vm.generateState();
    const hash = vmState.hash;

    let states = this.stateMap.get(vm.ic());
    if (!states) {
      states = new Set<string>();
      this.stateMap.set(vm.ic(), states);
    }

    if (states.has(hash)) {
      return true;
    }

    states.add(hash);
    this.totalStates++;
    return false;
  }
}
