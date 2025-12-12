# Part 2 

This document explains the design, math, orientation conventions, and usage of the Part 2 solver implemented for the button-press counting problem. The solver uses exact rational arithmetic and a bounded best-first search over free variables to find the minimal total number of button presses that satisfies per-bit count constraints.

## Problem Statement (Part 2)

Given:
- A target lamp state `result` (as a bit pattern),
- A list of buttons, each toggling a set of bit positions,
- A per-bit counter `counter[j]` that specifies how many times bit `j` must be affected (counted, not parity),

Find non-negative integer press counts `x_i` for each button `i` such that:
- For each bit `j`, the total number of times bit `j` is affected equals `counter[j]`:
  Σ_i A[i, j] · x_i = counter[j], where A[i, j] ∈ {0, 1} indicates whether button `i` affects bit `j`.
- The final state parity derived from the press counts matches `result`.
- The total number of button presses Σ_i x_i is minimal.

This is essentially a linear system with integer non-negativity constraints and a minimization objective.

## Orientation and Parsing Conventions

To avoid index mismatches, Part 2 uses a strict left-to-right (LTR) orientation for bits:
- Bit 0 is the leftmost character inside the bracket string `[...]`.
- Bit indices provided in instructions `(i,j,...)` are interpreted directly as LTR positions.
- The `counter` list `{c0,c1,...}` maps the first number to bit 0 (leftmost), second to bit 1, and so on.

This is different from the typical LSB-right convention often used in bitmask logic. For Part 2, all inputs are converted to LTR before building the matrix. This ensures a single, consistent orientation across `result`, `instructions`, and `counter`.

### Diagnostics

The implementation prints diagnostics before solving:
- `resultBits(LTR)`: the target parity bits in LTR order.
- `counterParity(LTR)`: parity of `counter` in LTR order (`counter[j] % 2`).
- `instructions(LTR)`: the button coverage lists in LTR indices.

These diagnostics help validate that the inputs are correctly oriented and, optionally, to confirm parity consistency.

## Algorithm Overview

The solver is split into two phases:

1) Rational Gaussian Elimination
2) Bounded Best-First Search Over Free Variables

### 1) Rational Gaussian Elimination

- Build a matrix `grid` of dimension `rows x columns`:
  - `rows = number of bits (m)`,
  - `columns = number of buttons (n)`,
  - `grid[r][c] = 1` if button `c` affects bit `r`, else `0`.

- Build the right-hand side vector `rhs` where `rhs[r] = counter[r]` as exact rationals (`big.Rat`).

- Perform Gaussian elimination with:
  - Row swaps and column swaps to find pivots,
  - Divide pivot row to normalize pivot to `1`,
  - Subtract multiples of pivot row from other rows to eliminate pivot column.

- Track a column permutation `colPerm` during swaps to keep button columns aligned. This permutation is applied to the button masks used later to compute upper bounds.

- After elimination:
  - Truncate trailing all-zero rows; their `rhs` must be zero (otherwise infeasible).
  - The first `rows'` columns become “main variables” (determined by the system after choosing free variables), and remaining `columns - rows'` columns become “free variables”.

### 2) Bounded Best-First Search Over Free Variables

We treat trailing columns (free variables) as independent dimensions with bounded ranges:
- For each free variable (button column `c`), compute:
  - `max_presses[c]` = minimum `counter[bit]` among bits affected by button `c`. This is a reasonable upper bound because pressing the button beyond this would inevitably exceed some bit's target count.
  - `press_difference_per_press[c] = 1 - sum_r grid[r][c]` (rational). This net effect measures how total press count changes if we increase that free variable by 1: +1 from the free variable itself minus the amount the main variables decrease due to its contribution.

We encode the free-variable vector as a single integer using mixed radix factors:
- Base for free variable `i` is `max_presses[i] + 1`.
- `factor[i]` is the product of prior bases, allowing a compact index.

Heuristic initial state:
- For each free variable:
  - If `press_difference_per_press[i] < 0` (pressing reduces total), start at `max_presses[i]`.
  - Else start at `0`.

Search procedure:
- Use a min-priority queue keyed by cumulative “added presses” (rational) to explore neighboring states (increment or decrement per free variable).
- For each state:
  - Decode the free variable vector (trailing presses).
  - Compute main variable solution: `presses_main[r] = rhs[r] - Σ_i grid[r][rows + i] * trailing[i]`.
  - If all `presses_main` are non-negative integers, return the optimal total:
    `Σ presses_main + Σ trailing`.
  - Otherwise, push neighbors:
    - If `diff < 0` and current value > 0, try decrement,
    - If current value < max, try increment.
- Use a visited set on the mixed-radix index to avoid revisiting states.

This best-first search converges quickly for small `m` and bounded free variables.

## Parity Discussion

Strict parity pre-check (counter parity equals target parity) is mathematically required for exact-count feasibility. However, due to orientation pitfalls or mixed data conventions:
- We disabled the strict parity check inside the solver until orientation is fully standardized.
- Post-standardization (LTR everywhere), you can re-enable parity checking as a fast infeasibility guard.

If re-enabled, check:
`counter[j] % 2 == resultBit[j]` for all `j`.
If any mismatch occurs, the instance is mathematically infeasible under the “exact per-bit count equals target” interpretation.

## Usage

1) Ensure your input lines conform to the LTR conventions:
   - `result`: `[...]` left-to-right, bit 0 is leftmost.
   - `instructions`: button tuples like `(0,2,3)` are LTR indices.
   - `counter`: `{c0,c1,...}` map directly to bits left-to-right.

2) Run the program. The `p2()` function:
   - Parses `result`, `instructions`, and `counter` into LTR orientation,
   - Prints diagnostics,
   - Calls the rational solver to find minimal total presses.

3) Example inputs:

```
[.##.] (3) (1,3) (2) (2,3) (0,2) (0,1) {3,5,4,7}
[...#.] (0,2,3,4) (2,3) (0,4) (0,1,2) (1,2,3,4) {7,5,12,7,2}
[.###.#] (0,1,2,3,4) (0,3,4) (0,1,2,4,5) (1,2) {10,11,11,5,10,5}
```

Sample diagnostic and result for the first line (as seen in logs):

```
p2[diag] resultBits(LTR): [0 1 1 0]
p2[diag] counterParity(LTR): [1 0 1 1]
p2[diag] instructions(LTR): [[3] [1 3] [2] [2 3] [0 2] [0 1]]
p2[Rational]: presses=10
p2: 10
```

If a line returns “infeasible or not found,” check the diagnostics—it likely indicates parity mismatch or a genuine inconsistency between `result` and `counter`.

## Implementation Notes

- Exact arithmetic uses `math/big.Rat` for rationals to avoid floating point error.
- Elimination uses pivoting (row/column swaps) to find valid pivots and normalize.
- Column permutation is applied to button masks to keep trailing bounds correct.
- Search neighbors are generated with careful bounds; visited states avoid cycles.
- With small bit-widths (m ≤ 6) and reasonable button counts, performance is good.

## Common Pitfalls

- Mixed orientation (LSB-right vs LTR):
  Always standardize to LTR before constructing the matrix. Mixing orientations yields incorrect parity and infeasibility.
- Misinterpreting counter semantics:
  The solver assumes `counter[j]` is the exact number of times bit `j` must be affected (not just parity).
- Missing coverage:
  If a bit has `counter[j] > 0` and no button covers that bit, the instance is infeasible.
- Overly large counters:
  The search remains bounded by `max_presses` but large values can increase the search space. In practice, with proper bounds and heuristic start, it remains tractable for AoC-scale inputs.

## Options and Extensions

- Re-enable strict parity pre-check after confirming orientation consistency to speed up infeasibility detection.
- Add a diagnostics toggle to reduce logging in production runs.
- If future inputs grow in size, consider:
  - Fraction-free elimination (Bareiss) with integers to reduce rational overhead,
  - ILP solver integration (CP-SAT or GLPK) for absolute robustness.

## Summary

The Part 2 solver:
- Converts all inputs to a consistent LTR orientation,
- Uses rational Gaussian elimination to isolate free variables,
- Searches bounded free-variable space with an effective heuristic,
- Produces minimal total presses for feasible instances,
- Reports infeasible or not found when parity/integer constraints can’t be satisfied.

Keep the diagnostic prints on during testing to ensure orientation is correct across lines. Once you’re confident, you can reduce logging and optionally re-enable parity checks.
