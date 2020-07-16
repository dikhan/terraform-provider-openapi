# Testing

- [Prerequisites](#prerequisites)
- [How to run all tests](#how-to-run-all-tests)
- [General guidelines](#general-guidelines)
- [Example tests](#example-tests)
  - [Single test case](#single-test-case)
  - [Multiple test cases](#multiple-test-cases)

## Prerequisites
- Install [goconvey](https://github.com/smartystreets/goconvey) by running `go get github.com/smartystreets/goconvey` 

## How to run all tests
The following make targets can be used to run all tests in this in project:
- `make test` will run all unit tests in addition to `gofmt`, `govet`, `golint`, and `gosec`
- `make test-all` will run `make test` in addition to all integration tests

## General guidelines
- Test name should be `Test` plus the method name tested (eg: `TestIsBoolExtensionEnabled`) 
- Inputs parameters and expected outputs should be identifiable by their names
  - Inputs should have names beginning with `input` (eg: `inputStatusCode`)
  - Expected outputs should have names beginning with `expected` (eg: `expectedResult`)
- Organize tests using `Given`, `When`, and `Then` [Convey](https://github.com/smartystreets/goconvey/wiki#your-first-goconvey-test) statements to make the intention of the test clear:
  - `Given`: The context/setup for the behavior being tested
  - `When`: The behavior being tested
  - `Then`: Assert on the expected outputs
- Assertions should use the relevant [So](https://github.com/smartystreets/goconvey/wiki/Assertions) comparison
- Tests with multiple test cases should follow a combined table driven test and Convey format. See the [multiple tests cases example](#multiple-test-cases).

## Example tests
### Single test case
No need to follow the table driven test format for one test case.
```
func TestNewSpecV2Resource(t *testing.T) {
	Convey("Given a root path /users, a root path item, schema definitions", t, func() {
		inputPath := "/users"
		inputRootPathItem := spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Post: &spec.Operation{},
			},
		}
		inputSchemaDefinitions := map[string]spec.Schema{}
		Convey("When the newSpecV2Resource method is called", func() {
			r, err := newSpecV2Resource(inputPath, spec.Schema{}, inputRootPathItem, spec.PathItem{}, inputSchemaDefinitions, map[string]spec.PathItem{})
			Convey("Then the resource returned should have name `users` and there should be no error", func() {
				So(r.GetResourceName(), ShouldEqual, "users")
				So(err, ShouldBeNil)
			})
		})
	})
}
```
### Multiple test cases
Follow the table driven test format with a Convey statement wrapper around the test cases loop.

The `When` Convey statement should include a reference to the test case name. This allows you to differentiate between test cases when running tests in verbose mode (`go test -v`) since each test case name will be printed.
```
func TestIsBoolExtensionEnabled(t *testing.T) {
	testCases := []struct {
		name           string
		inputExtensions     spec.Extensions
		inputExtension      string
		expectedResult bool
		expectedError error
	}{
		{name: "nil extensions", inputExtensions: nil, inputExtension: "", expectedResult: false},
		{name: "empty extensions", inputExtensions: spec.Extensions{}, inputExtension: "", expectedResult: false},
		{name: "populated extensions but empty extension", inputExtensions: spec.Extensions{"some-extension": true}, inputExtension: "", expectedResult: false},
		{name: "populated extensions and matching bool extension with value true", inputExtensions: spec.Extensions{"some-extension": true}, inputExtension: "some-extension", expectedResult: true},
		{name: "populated extensions and matching bool extension with value false", inputExtensions: spec.Extensions{"some-extension": false}, inputExtension: "some-extension", expectedResult: false},
		{name: "populated extensions and matching non bool extension", inputExtensions: spec.Extensions{"some-extension": "some value"}, inputExtension: "some-extension", expectedResult: false},
	}
	Convey("Given a SpecV2Resource", t, func() {
		r := &SpecV2Resource{}
		for _, tc := range testCases {
			Convey(fmt.Sprintf("When isBoolExtensionEnabled method is called: %s", tc.name), func() {
				isEnabled, err := r.testMethod(tc.inputExtensions, tc.inputExtension)
				Convey("Then the result returned should be the expected one", func() {
					So(err, ShouldResemble, tc.expectedError)
					So(isEnabled, ShouldEqual, tc.expectedResult)
				})
			})
		}
	})
}
```

### Methods that don't require setup (Given)

In some cases, the method under tests may not need any special set up. The following example can be used as an example on
how to write such tests:

````
func TestExpandPath(t *testing.T) {

	testCases := []struct {
		name           string
		inputPath     string
		expectedResult string
		expectedError error
	}{
		{name: "input example 1", inputPath: "input1", expectedResult: "result1", expectedError: nil,},
		{name: "input example 2", inputPath: "input2",  expectedResult: "result2", expectedError: nil,},
	}

	for _, tc := range testCases {
		Convey(fmt.Sprintf("When testMethod method is called: %s", tc.name), t, func() {
			returnedPath, err := testMethod(tc.inputPath)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldResemble, tc.expectedError)
				So(returnedPath, ShouldEqual, tc.expectedResult)
			})
		})
	}
}
````