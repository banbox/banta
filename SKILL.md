---
name: banta-indicator
description: Implement and validate a Banta technical indicator with matching stateful and vector APIs.
---

# Banta indicator implementation

Use this skill when adding one technical indicator to `/data/ban/banta`. The
implementation is complete only when the streaming (`banta`) and vector
(`banta/tav`) versions have the same documented semantics, the result has been
cross-checked against other TA libraries available to the current Python TA
environment, and deterministic correctness and performance tests pass.

## Repository map

- Stateful indicators: `indicators_*.go` (stateful package) (`package banta`).
- Vector indicators: `tav/indicators_*.go` (vector package) (`package tav`).
- Stateful/vector comparison helpers: `core_test.go`, `base_test.go`,
  `test_helpers_test.go`, and the family-focused `indicators_*_test.go` files.
- Go module: `/data/ban/banta/go.mod` (`github.com/banbox/banta`).
- Python-facing wrappers (only when the public binding is required):
  `python/ta/index.go` and `python/tav/index.go`.
- Cross-library checker: **always use `/data/ban/banta/scripts/test_inds.py`**.
  This is the repository's authoritative comparison harness. It runs the
  same OHLCV fixture through TA-Lib, pandas-ta, MyTT and other configured
  Python indicator libraries. Do not replace this with tests under
  `python/`, and do not treat building or importing the gopy wrapper as a
  cross-library numerical comparison.

Before editing, search the matching stateful and vector family files for the
indicator and its aliases. Reuse an existing primitive or add a compatibility
alias when the formula already exists. Do not silently map an unsupported
indicator to a different formula.

## Required workflow

### 1. Establish the contract

Record the indicator name, source aliases, inputs, parameters, output count,
warm-up period, smoothing method, and behavior for invalid or missing values.
Use the call sites found in `origin/` to identify parameter conventions; a
name such as `ATR`, `MFI`, or `STOCH` can have different defaults in different
libraries.

Choose a reference implementation before writing Go code. Prefer a pinned
version of `pandas_ta` or TA-Lib for numerical indicators, and use the
TradingView/Pine definition when the source strategy explicitly uses Pine
semantics. Keep the reference source path, version, parameters, and any
adaptation in the indicator's test or design note. Compare variants rather
than assuming that similarly named functions are equivalent. Resolve at
least:

- SMA/EMA/Wilder/RMA initialization and the first valid output;
- rolling-window inclusion, ties, offsets, and lookback indexing;
- `NaN` at the beginning and in the middle of a series;
- standard deviation degrees of freedom and zero denominators;
- session/reset behavior for cumulative calculations;
- output ordering and whether a result is a scalar, flag, or multiple lines.

### 1a. Cross-library discovery and oracle selection (required)

Before finalizing the formula, add a `test_<indicator>()` function to
`/data/ban/banta/scripts/test_inds.py` (or extend its explicit dispatch) and
run that function. The test must inspect every installed or locally available
technical-indicator library supported by that TA workspace,
record each library and version, and report whether an actually callable
same-name implementation exists. A documentation name match is insufficient:
import the function and run it on the shared deterministic fixture with the
exact parameters and input columns.

When several experimental implementations exist, select the variant used by
the most strategies or call sites in the current TA corpus. Record each usage
count and this selection rule. If exactly one library implementation exists,
verify that its formula and parameter semantics also match the strategy source
that motivated the indicator. If no implementation exists, record that fact
and use the pinned strategy/reference definition instead.

The script must compare every output line and every bar, including warm-up and
missing-value positions. Normalize only representation-level differences (for
example `-0.0` and NaN payloads). Cross-language floating-point differences
may use the explicit bound required by section 3b, but never an undocumented
tolerance. Any intentional adaptation (reset, seed, or lookback)
must document the reason and first differing bar, and have a focused assertion.
Save the fixture, selected-oracle metadata, and machine-readable comparison
result so the check reruns without network access. The completion note must
include the exact command, for example:

```bash
cd /data/ban/banta
python3 scripts/test_inds.py --indicator <name>
```

If the current script has no dispatch option, add one while preserving its
existing default behavior (`test_mfi`). The new dispatch must call the same
`print_tares`/library calculations as the manually callable test function.

### 2. Implement both APIs

Add the stateful API to `indicators_*.go`. Follow the existing pattern:

```go
func Example(obj *Series, period int) *Series {
	res := obj.To("_example", period)
	if res.Cached() {
		return res
	}
	// Keep mutable rolling state in res.More. Append exactly one value per bar.
	res.Append(obj.Get(0))
	return res
}
```

For stateful code:

- Keep mutable rolling sums, windows, and prior values in a private state type
  stored in `res.More`.
- Set `res.DupMore` and deep-copy slices so cloned environments cannot share
  mutable state.
- Use `obj.To` with a unique function key and all parameters represented in the
  cache key. Match existing dependency composition where possible.
- Append one output for every input bar, including `math.NaN()` during warm-up
  or when the contract says the value is undefined.
- Treat isolated `NaN` values explicitly. Follow the selected reference's
  behavior for skipping an individual missing observation, and preserve `NaN`
  where the reference says output is undefined. Do not let one missing input
  poison every later value, turn into zero, or silently alter the window
  length. Cover leading and interior `NaN` values in fixtures and tests.
- Return multiple lines in a stable order and document that order in a comment.

Add the vector API to `tav/indicators_*.go` with the same parameter order and
output order:

```go
func Example(data []float64, period int) []float64 {
	result := make([]float64, len(data))
	// Compute without reading state from the banta package.
	return result
}
```

The vector function must return one value per input bar (or one equal-length
slice per output line), preserve the agreed `NaN` and warm-up behavior, and
avoid mutating caller-owned input slices. If a vector implementation uses a
helper, verify that helper has the same reset, initialization, and isolated
`NaN` skipping semantics as the stateful implementation.

After both Go APIs are stable, expose the indicator through both Python
wrappers in `python/ta/index.go` and `python/tav/index.go` only if the
indicator is part of the public binding contract. Wrapper execution is a
separate binding smoke test; it never replaces the `scripts/test_inds.py`
cross-library comparison.

### 3. Add independent fixtures and tests

Prefer a dedicated test file such as `indicator_<name>_test.go` to reduce
merge conflicts between parallel implementations. Keep a small deterministic
OHLCV fixture in `testdata/` or as a named fixture in the test file. Expected
values must come from the pinned reference implementation or hand-calculated
examples, never from either Banta implementation under test. This fixture is
the independent oracle; do not derive expected vector values from the
stateful result or expected stateful values from the vector result.

Every indicator test should cover:

1. ordinary values after warm-up;
2. the first valid output and all warm-up positions;
3. an initial `NaN` and a middle `NaN` when the contract permits missing data;
4. boundary parameters (period 1, the shortest valid period, and invalid
   period handling);
5. zero volume or zero denominator for volume/ratio indicators;
6. ties and flat prices for extrema and directional indicators;
7. every output line for multi-output indicators.

Use `RunFakeEnv` to feed `Kline` values into the stateful function. This is a
per-bar streaming unit-test harness; it does not compile or execute a trading
strategy and says nothing about profitability. Use the existing `CaseItem`
shape and `runAndCompareCases` when it fits; otherwise add a focused helper in
the dedicated test file. Assert lengths, output-line ordering, and every bar
against the independent oracle.

Compare the two Banta Go APIs bar by bar. Exact equality is preferred; allow
only the same explicit floating-point bound used for the external comparison
(default absolute/relative `1e-8`) when the two implementations perform the
same formula in a different arithmetic order. Treat two NaNs as equal and
record the first difference and maximum error. Any warm-up, indexing, or
algorithmic mismatch is a failure and must be fixed rather than hidden by the
tolerance.

Run the same fixture through the selected external Python oracle with the
exact parameters, and save its expected values or a reproducible generator.
Compare every output line and bar using the strict rule in section 1a; do not
weaken checks with tolerances. If the selected library differs because of
initialization, warm-up, missing-value, or reset semantics, compare the other
available variants and select the strategy-compatible definition. Record the
reason, first differing bar, and a focused assertion for any deliberate
adaptation. If no callable reference exists, use a pinned source definition
and independently hand-calculated expected values. Such an adaptation does
not waive exact parity between Banta's own APIs.

### 3a. Run the repository Python cross-library comparison (required)

Run `/data/ban/banta/scripts/test_inds.py`, not the Go-to-Python binding
tests. The new indicator's test function must execute the indicator in every
callable reference library available in that environment and print all output
lines with the existing `print_tares` helper. For strict machine comparison,
add a small JSON result writer or a companion assertion in the same script;
the expected values must come from the external libraries, never from Banta.
Record library names and versions, exact parameters, first differing bar,
warm-up/NaN behavior, and the selected variant. A wrapper smoke test may be
run separately, but it is optional and cannot mark an indicator cross-checked.

### 3b. Compare the external result with both Go APIs (required)

The Python output is only an oracle; merely running `test_inds.py` is
insufficient. Use the exact same fixture and parameters emitted by
`test_inds.py --json` as input to a Go comparison runner. The runner must
execute the vector function and replay the same candles through the stateful
function with `RunFakeEnv`. Compare every output line and every bar from both
Go paths to the selected Python library output. Treat two NaNs as equal and
allow only a small, explicit floating-point error (default relative/absolute
bound `1e-8`) for either Go path; report the first differing bar and maximum
absolute/relative error for state/vector and oracle comparisons separately.

Save one machine-readable artifact containing the fixture hash, indicator,
parameters, Python library versions, Go vector output, Go stateful output,
selected oracle output, comparison tolerances, first difference, and pass/fail
status. An indicator is accepted only when both Go paths pass this same-fixture
comparison, or when a documented library-specific warm-up/seed adaptation is
covered by a focused regression assertion.

### 4. Validate and report

From the Banta checkout, format and run the focused tests first:

```bash
cd /data/ban/banta
go fmt . ./tav
go test . -run 'Test<IndicatorName>'
go test ./tav -run 'Test<IndicatorName>'
```

Then run the package suites:

```bash
go test ./tav
go test ./...
```

Keep strategy profitability backtests separate from indicator tests. A
`RunFakeEnv` replay proves per-bar calculation behavior only. A strategy
backtest means compiling and running a strategy through the project's actual
backtest engine and inspecting its orders and performance evidence. Run that
additional check when the task requires strategy-level regression or
profitability evidence; never describe a unit replay as a profitability
backtest. A strategy producing orders also does not prove the indicator's
stateful, vector, or Python APIs are numerically correct.

If an unrelated existing test fails, preserve its output and report the
failure separately; do not change unrelated fixtures to make the new
indicator pass. Check the diff and verify that the implementation does not
introduce a duplicate exported name:

```bash
git diff --check
git diff -- indicators_*.go tav/indicators_*.go
```

The completion note for each indicator should include the reference library
and version, all same-name libraries discovered, usage counts, the selected
oracle and why it won, the exact `/data/ban/banta/scripts/test_inds.py
--indicator <name> --json <artifact>` command and artifact, the chosen formula
variant, state/vector function names, fixture location, focused Go test,
package test result, and any known numerical or `NaN` differences. If public
gopy wrappers were also changed, report their smoke test separately; it is not
the cross-library evidence.
Stateful Go, vector Go, and Python-facing results are validated only after
all applicable APIs have actually executed and matched the independent oracle
and each other under the strict comparison rule.

## Parallel implementation rules

When several agents implement indicators, give each agent one indicator and a
dedicated test file or fixture directory. Agents must not edit shared test
helpers or rewrite another indicator's state. The integrating agent resolves
export-name collisions, runs the full suite, and checks that all selected
indicators meet the minimum strategy-reference count before accepting them.
