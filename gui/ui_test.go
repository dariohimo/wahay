package gui

import (
	"errors"
	"os"
	"testing"

	. "gopkg.in/check.v1"

	"github.com/coyim/gotk3adapter/glibi"
	"github.com/coyim/gotk3adapter/gtk_mock"
	"github.com/coyim/gotk3adapter/gtki"
)

func Test(t *testing.T) { TestingT(t) }

type TonioGUISuite struct{}

var _ = Suite(&TonioGUISuite{})

type testGtkStruct struct {
	gtk_mock.Mock

	initCalled bool
	initArgs   *[]string

	applicationNewToReturn1 gtki.Application
	applicationNewToReturn2 error
	applicationNewCalled    bool
	applicationNewArg1      string
	applicationNewArg2      glibi.ApplicationFlags
}

func (t *testGtkStruct) Init(a *[]string) {
	t.initCalled = true
	t.initArgs = a
}

func (t *testGtkStruct) ApplicationNew(a1 string, a2 glibi.ApplicationFlags) (gtki.Application, error) {
	t.applicationNewCalled = true
	t.applicationNewArg1 = a1
	t.applicationNewArg2 = a2

	return t.applicationNewToReturn1, t.applicationNewToReturn2
}

func (s *TonioGUISuite) Test_CreateGraphics_createsAGraphicsObjectWithTheArgumentProvided(c *C) {
	g1 := CreateGraphics(nil, nil)
	c.Assert(g1.gtk, IsNil)
	c.Assert(g1.glib, IsNil)

	v := &testGtkStruct{}
	g2 := CreateGraphics(v, nil)
	c.Assert(g2.gtk, Equals, v)
	c.Assert(g2.glib, IsNil)
}

func (s *TonioGUISuite) Test_argsWithApplicationName(c *C) {
	org := os.Args
	os.Args = []string{"one", "two", "four"}
	defer func() {
		os.Args = org
	}()

	c.Assert(*argsWithApplicationName(), DeepEquals, []string{"Tonio", "two", "four"})

	os.Args[2] = "something"

	c.Assert(*argsWithApplicationName(), DeepEquals, []string{"Tonio", "two", "something"})
}

func (s *TonioGUISuite) Test_NewGTK_initializesGTK(c *C) {
	org := os.Args
	os.Args = []string{"a", "b", "cee"}
	defer func() {
		os.Args = org
	}()

	ourGtk := &testGtkStruct{}
	g1 := CreateGraphics(ourGtk, nil)
	_ = NewGTK(g1)

	c.Assert(ourGtk.initCalled, Equals, true)
	c.Assert(*ourGtk.initArgs, DeepEquals, []string{"Tonio", "b", "cee"})
}

func (s *TonioGUISuite) Test_NewGTK_callsApplicationNewWithAppropriateArguments(c *C) {
	ourGtk := &testGtkStruct{}
	g1 := CreateGraphics(ourGtk, nil)
	_ = NewGTK(g1)

	c.Assert(ourGtk.applicationNewCalled, Equals, true)
	c.Assert(ourGtk.applicationNewArg1, Equals, "digital.autonomia.Tonio")
	c.Assert(ourGtk.applicationNewArg2, Equals, glibi.APPLICATION_FLAGS_NONE)
}

func (s *TonioGUISuite) Test_NewGTK_panicsIfApplicationNewFails(c *C) {
	ourGtk := &testGtkStruct{}
	ourGtk.applicationNewToReturn2 = errors.New("something went wrong")
	g1 := CreateGraphics(ourGtk, nil)

	c.Assert(func() { NewGTK(g1) }, PanicMatches, "Couldn't create application: something went wrong")
}

func (s *TonioGUISuite) Test_NewGTK_returnsAGTKUIWithProperData(c *C) {
	app := &gtk_mock.MockApplication{}
	ourGtk := &testGtkStruct{}
	ourGtk.applicationNewToReturn1 = app
	g1 := CreateGraphics(ourGtk, nil)
	ret := NewGTK(g1).(*gtkUI)

	c.Assert(ret.app, Equals, app)
	c.Assert(ret.g, Equals, g1)
}

func (s *TonioGUISuite) Test_gtkUI_onActivate_createsMainWindow(c *C) {
	ourGtk := &testGtkWithBuilder{}
	ourBuilder := &testBuilder{}
	ourGtk.builderNewToReturn1 = ourBuilder
	ourAppWindow := &gtk_mock.MockApplicationWindow{}
	ourBuilder.getObjectToReturn1 = ourAppWindow

	g1 := CreateGraphics(ourGtk, nil)

	u := &gtkUI{g: g1}
	u.onActivate()

	c.Assert(ourBuilder.getObjectArg1, Equals, "mainWindow")
}

type testApplication struct {
	gtk_mock.MockApplication

	connectArg1    string
	connectArg2    interface{}
	connectArgRest []interface{}
	connectReturn1 glibi.SignalHandle
	connectReturn2 error

	runArg1    []string
	runReturn1 int
}

func (ta *testApplication) Connect(v1 string, v2 interface{}, v3 ...interface{}) (glibi.SignalHandle, error) {
	ta.connectArg1 = v1
	ta.connectArg2 = v2
	ta.connectArgRest = v3
	return ta.connectReturn1, ta.connectReturn2
}

func (ta *testApplication) Run(v1 []string) int {
	ta.runArg1 = v1
	return ta.runReturn1
}

func (s *TonioGUISuite) Test_gtkUI_Loop_connectsActivate(c *C) {
	app := &testApplication{}
	u := &gtkUI{app: app}
	u.Loop()

	c.Assert(app.connectArg1, Equals, "activate")
	c.Assert(app.connectArg2, Not(IsNil))
}

func (s *TonioGUISuite) Test_gtkUI_Loop_runsTheAppWithArguments(c *C) {
	app := &testApplication{}
	u := &gtkUI{app: app}
	u.Loop()

	c.Assert(app.runArg1, DeepEquals, []string{})
}

func (s *TonioGUISuite) Test_gtkUI_Loop_panicsIfActivateFails(c *C) {
	app := &testApplication{}
	u := &gtkUI{app: app}
	app.connectReturn2 = errors.New("oh nooo")
	c.Assert(func() { u.Loop() }, PanicMatches, "Couldn't activate application: oh nooo")
}
