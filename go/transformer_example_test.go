// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestTransformer_TransformerInFunc_TransformIn_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject TransformerInFunc[any, any]
		_, _ = subject.TransformIn(nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestTransformer_TransformerInFunc_TransformIn_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject TransformerInFunc[any, any]
		_, _ = subject.TransformIn(nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestTransformer_TransformerInFunc_TransformIn_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject TransformerInFunc[any, any]
		_, _ = subject.TransformIn(nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestTransformer_TransformerOutFunc_TransformOut_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject TransformerOutFunc[any, any]
		_, _ = subject.TransformOut(nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestTransformer_TransformerOutFunc_TransformOut_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject TransformerOutFunc[any, any]
		_, _ = subject.TransformOut(nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestTransformer_TransformerOutFunc_TransformOut_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject TransformerOutFunc[any, any]
		_, _ = subject.TransformOut(nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestTransformer_RenameFields_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = RenameFields(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestTransformer_RenameFields_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = RenameFields(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestTransformer_RenameFields_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = RenameFields(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestTransformer_FieldRenamer_TransformIn_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject FieldRenamer
		_, _ = subject.TransformIn(nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestTransformer_FieldRenamer_TransformIn_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject FieldRenamer
		_, _ = subject.TransformIn(nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestTransformer_FieldRenamer_TransformIn_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject FieldRenamer
		_, _ = subject.TransformIn(nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestTransformer_FieldRenamer_TransformOut_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject FieldRenamer
		_, _ = subject.TransformOut(nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestTransformer_FieldRenamer_TransformOut_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject FieldRenamer
		_, _ = subject.TransformOut(nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestTransformer_FieldRenamer_TransformOut_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject FieldRenamer
		_, _ = subject.TransformOut(nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleTransformerInFunc_TransformIn_transformer() {
	func() {
		defer func() { _ = recover() }()
		var subject TransformerInFunc[any, any]
		_, _ = subject.TransformIn(nil, nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleTransformerOutFunc_TransformOut_transformer() {
	func() {
		defer func() { _ = recover() }()
		var subject TransformerOutFunc[any, any]
		_, _ = subject.TransformOut(nil, nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRenameFields_transformer() {
	func() {
		defer func() { _ = recover() }()
		_ = RenameFields(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleFieldRenamer_TransformIn_transformer() {
	func() {
		defer func() { _ = recover() }()
		var subject FieldRenamer
		_, _ = subject.TransformIn(nil, nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleFieldRenamer_TransformOut_transformer() {
	func() {
		defer func() { _ = recover() }()
		var subject FieldRenamer
		_, _ = subject.TransformOut(nil, nil)
	}()
	coretest.Println("done")
	// Output: done
}
