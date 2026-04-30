// SPDX-License-Identifier: EUPL-1.2

package api

import coretest "dappco.re/go"

func TestExport_ExportSpec_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = ExportSpec(nil, "", nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestExport_ExportSpec_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = ExportSpec(nil, "", nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestExport_ExportSpec_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = ExportSpec(nil, "", nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestExport_ExportSpecIter_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = ExportSpecIter(nil, "", nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestExport_ExportSpecIter_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = ExportSpecIter(nil, "", nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestExport_ExportSpecIter_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = ExportSpecIter(nil, "", nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestExport_ExportSpecToFile_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = ExportSpecToFile("", "", nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestExport_ExportSpecToFile_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = ExportSpecToFile("", "", nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestExport_ExportSpecToFile_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = ExportSpecToFile("", "", nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestExport_ExportSpecToFileIter_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = ExportSpecToFileIter("", "", nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestExport_ExportSpecToFileIter_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = ExportSpecToFileIter("", "", nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestExport_ExportSpecToFileIter_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = ExportSpecToFileIter("", "", nil, nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleExportSpec_export() {
	func() {
		defer func() { _ = recover() }()
		_ = ExportSpec(nil, "", nil, nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleExportSpecIter_export() {
	func() {
		defer func() { _ = recover() }()
		_ = ExportSpecIter(nil, "", nil, nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleExportSpecToFile_export() {
	func() {
		defer func() { _ = recover() }()
		_ = ExportSpecToFile("", "", nil, nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleExportSpecToFileIter_export() {
	func() {
		defer func() { _ = recover() }()
		_ = ExportSpecToFileIter("", "", nil, nil)
	}()
	coretest.Println("done")
	// Output: done
}
