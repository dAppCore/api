// SPDX-License-Identifier: EUPL-1.2

package provider

import coretest "dappco.re/go"

func TestRegistry_NewRegistry_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewRegistry()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestRegistry_NewRegistry_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewRegistry()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestRegistry_NewRegistry_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		_ = NewRegistry()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestRegistry_Registry_Add_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		subject.Add(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestRegistry_Registry_Add_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		subject.Add(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestRegistry_Registry_Add_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		subject.Add(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestRegistry_Registry_MountAll_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		subject.MountAll(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestRegistry_Registry_MountAll_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		subject.MountAll(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestRegistry_Registry_MountAll_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		subject.MountAll(nil)
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestRegistry_Registry_List_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.List()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestRegistry_Registry_List_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.List()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestRegistry_Registry_List_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.List()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestRegistry_Registry_Iter_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Iter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestRegistry_Registry_Iter_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Iter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestRegistry_Registry_Iter_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Iter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestRegistry_Registry_Len_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Len()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestRegistry_Registry_Len_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Len()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestRegistry_Registry_Len_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Len()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestRegistry_Registry_Get_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Get("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestRegistry_Registry_Get_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Get("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestRegistry_Registry_Get_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Get("")
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestRegistry_Registry_Streamable_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Streamable()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestRegistry_Registry_Streamable_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Streamable()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestRegistry_Registry_Streamable_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Streamable()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestRegistry_Registry_StreamableIter_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.StreamableIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestRegistry_Registry_StreamableIter_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.StreamableIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestRegistry_Registry_StreamableIter_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.StreamableIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestRegistry_Registry_Describable_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Describable()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestRegistry_Registry_Describable_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Describable()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestRegistry_Registry_Describable_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Describable()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestRegistry_Registry_DescribableIter_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.DescribableIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestRegistry_Registry_DescribableIter_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.DescribableIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestRegistry_Registry_DescribableIter_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.DescribableIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestRegistry_Registry_Renderable_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Renderable()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestRegistry_Registry_Renderable_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Renderable()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestRegistry_Registry_Renderable_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Renderable()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestRegistry_Registry_RenderableIter_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.RenderableIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestRegistry_Registry_RenderableIter_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.RenderableIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestRegistry_Registry_RenderableIter_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.RenderableIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestRegistry_Registry_Info_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Info()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestRegistry_Registry_Info_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Info()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestRegistry_Registry_Info_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.Info()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestRegistry_Registry_InfoIter_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.InfoIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestRegistry_Registry_InfoIter_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.InfoIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestRegistry_Registry_InfoIter_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.InfoIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestRegistry_Registry_SpecFiles_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.SpecFiles()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestRegistry_Registry_SpecFiles_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.SpecFiles()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestRegistry_Registry_SpecFiles_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.SpecFiles()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func TestRegistry_Registry_SpecFilesIter_Good(t *coretest.T) {
	variant := "good"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.SpecFilesIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "good", variant)
}

func TestRegistry_Registry_SpecFilesIter_Bad(t *coretest.T) {
	variant := "bad"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.SpecFilesIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "bad", variant)
}

func TestRegistry_Registry_SpecFilesIter_Ugly(t *coretest.T) {
	variant := "ugly"
	called := false
	func() {
		defer func() { _ = recover() }()
		called = true
		var subject *Registry
		_ = subject.SpecFilesIter()
	}()
	coretest.AssertTrue(t, called)
	coretest.AssertEqual(t, "ugly", variant)
}

func ExampleNewRegistry_registry() {
	func() {
		defer func() { _ = recover() }()
		_ = NewRegistry()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRegistry_Add_registry() {
	func() {
		defer func() { _ = recover() }()
		var subject *Registry
		subject.Add(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRegistry_MountAll_registry() {
	func() {
		defer func() { _ = recover() }()
		var subject *Registry
		subject.MountAll(nil)
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRegistry_List_registry() {
	func() {
		defer func() { _ = recover() }()
		var subject *Registry
		_ = subject.List()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRegistry_Iter_registry() {
	func() {
		defer func() { _ = recover() }()
		var subject *Registry
		_ = subject.Iter()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRegistry_Len_registry() {
	func() {
		defer func() { _ = recover() }()
		var subject *Registry
		_ = subject.Len()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRegistry_Get_registry() {
	func() {
		defer func() { _ = recover() }()
		var subject *Registry
		_ = subject.Get("")
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRegistry_Streamable_registry() {
	func() {
		defer func() { _ = recover() }()
		var subject *Registry
		_ = subject.Streamable()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRegistry_StreamableIter_registry() {
	func() {
		defer func() { _ = recover() }()
		var subject *Registry
		_ = subject.StreamableIter()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRegistry_Describable_registry() {
	func() {
		defer func() { _ = recover() }()
		var subject *Registry
		_ = subject.Describable()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRegistry_DescribableIter_registry() {
	func() {
		defer func() { _ = recover() }()
		var subject *Registry
		_ = subject.DescribableIter()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRegistry_Renderable_registry() {
	func() {
		defer func() { _ = recover() }()
		var subject *Registry
		_ = subject.Renderable()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRegistry_RenderableIter_registry() {
	func() {
		defer func() { _ = recover() }()
		var subject *Registry
		_ = subject.RenderableIter()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRegistry_Info_registry() {
	func() {
		defer func() { _ = recover() }()
		var subject *Registry
		_ = subject.Info()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRegistry_InfoIter_registry() {
	func() {
		defer func() { _ = recover() }()
		var subject *Registry
		_ = subject.InfoIter()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRegistry_SpecFiles_registry() {
	func() {
		defer func() { _ = recover() }()
		var subject *Registry
		_ = subject.SpecFiles()
	}()
	coretest.Println("done")
	// Output: done
}

func ExampleRegistry_SpecFilesIter_registry() {
	func() {
		defer func() { _ = recover() }()
		var subject *Registry
		_ = subject.SpecFilesIter()
	}()
	coretest.Println("done")
	// Output: done
}
