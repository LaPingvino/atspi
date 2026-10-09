// Code generated from at-spi2-core's atspi/atspi-constants.h (main). DO NOT EDIT.

package atspi

// Role is what an object is (AtspiRole).
type Role uint32

// The roles.
const (
	RoleInvalid              Role = 0
	RoleAcceleratorLabel     Role = 1
	RoleAlert                Role = 2
	RoleAnimation            Role = 3
	RoleArrow                Role = 4
	RoleCalendar             Role = 5
	RoleCanvas               Role = 6
	RoleCheckBox             Role = 7
	RoleCheckMenuItem        Role = 8
	RoleColorChooser         Role = 9
	RoleColumnHeader         Role = 10
	RoleComboBox             Role = 11
	RoleDateEditor           Role = 12
	RoleDesktopIcon          Role = 13
	RoleDesktopFrame         Role = 14
	RoleDial                 Role = 15
	RoleDialog               Role = 16
	RoleDirectoryPane        Role = 17
	RoleDrawingArea          Role = 18
	RoleFileChooser          Role = 19
	RoleFiller               Role = 20
	RoleFocusTraversable     Role = 21
	RoleFontChooser          Role = 22
	RoleFrame                Role = 23
	RoleGlassPane            Role = 24
	RoleHtmlContainer        Role = 25
	RoleIcon                 Role = 26
	RoleImage                Role = 27
	RoleInternalFrame        Role = 28
	RoleLabel                Role = 29
	RoleLayeredPane          Role = 30
	RoleList                 Role = 31
	RoleListItem             Role = 32
	RoleMenu                 Role = 33
	RoleMenuBar              Role = 34
	RoleMenuItem             Role = 35
	RoleOptionPane           Role = 36
	RolePageTab              Role = 37
	RolePageTabList          Role = 38
	RolePanel                Role = 39
	RolePasswordText         Role = 40
	RolePopupMenu            Role = 41
	RoleProgressBar          Role = 42
	RoleButton               Role = 43
	RoleRadioButton          Role = 44
	RoleRadioMenuItem        Role = 45
	RoleRootPane             Role = 46
	RoleRowHeader            Role = 47
	RoleScrollBar            Role = 48
	RoleScrollPane           Role = 49
	RoleSeparator            Role = 50
	RoleSlider               Role = 51
	RoleSpinButton           Role = 52
	RoleSplitPane            Role = 53
	RoleStatusBar            Role = 54
	RoleTable                Role = 55
	RoleTableCell            Role = 56
	RoleTableColumnHeader    Role = 57
	RoleTableRowHeader       Role = 58
	RoleTearoffMenuItem      Role = 59
	RoleTerminal             Role = 60
	RoleText                 Role = 61
	RoleToggleButton         Role = 62
	RoleToolBar              Role = 63
	RoleToolTip              Role = 64
	RoleTree                 Role = 65
	RoleTreeTable            Role = 66
	RoleUnknown              Role = 67
	RoleViewport             Role = 68
	RoleWindow               Role = 69
	RoleExtended             Role = 70
	RoleHeader               Role = 71
	RoleFooter               Role = 72
	RoleParagraph            Role = 73
	RoleRuler                Role = 74
	RoleApplication          Role = 75
	RoleAutocomplete         Role = 76
	RoleEditbar              Role = 77
	RoleEmbedded             Role = 78
	RoleEntry                Role = 79
	RoleChart                Role = 80
	RoleCaption              Role = 81
	RoleDocumentFrame        Role = 82
	RoleHeading              Role = 83
	RolePage                 Role = 84
	RoleSection              Role = 85
	RoleRedundantObject      Role = 86
	RoleForm                 Role = 87
	RoleLink                 Role = 88
	RoleInputMethodWindow    Role = 89
	RoleTableRow             Role = 90
	RoleTreeItem             Role = 91
	RoleDocumentSpreadsheet  Role = 92
	RoleDocumentPresentation Role = 93
	RoleDocumentText         Role = 94
	RoleDocumentWeb          Role = 95
	RoleDocumentEmail        Role = 96
	RoleComment              Role = 97
	RoleListBox              Role = 98
	RoleGrouping             Role = 99
	RoleImageMap             Role = 100
	RoleNotification         Role = 101
	RoleInfoBar              Role = 102
	RoleLevelBar             Role = 103
	RoleTitleBar             Role = 104
	RoleBlockQuote           Role = 105
	RoleAudio                Role = 106
	RoleVideo                Role = 107
	RoleDefinition           Role = 108
	RoleArticle              Role = 109
	RoleLandmark             Role = 110
	RoleLog                  Role = 111
	RoleMarquee              Role = 112
	RoleMath                 Role = 113
	RoleRating               Role = 114
	RoleTimer                Role = 115
	RoleStatic               Role = 116
	RoleMathFraction         Role = 117
	RoleMathRoot             Role = 118
	RoleSubscript            Role = 119
	RoleSuperscript          Role = 120
	RoleDescriptionList      Role = 121
	RoleDescriptionTerm      Role = 122
	RoleDescriptionValue     Role = 123
	RoleFootnote             Role = 124
	RoleContentDeletion      Role = 125
	RoleContentInsertion     Role = 126
	RoleMark                 Role = 127
	RoleSuggestion           Role = 128
	RolePushButtonMenu       Role = 129
	RoleSwitch               Role = 130
	RolePushButton           Role = 43
)

// roleNames are the roles' names, as GetRoleName reports them.
var roleNames = map[Role]string{
	RoleInvalid:              "invalid",
	RoleAcceleratorLabel:     "accelerator label",
	RoleAlert:                "alert",
	RoleAnimation:            "animation",
	RoleArrow:                "arrow",
	RoleCalendar:             "calendar",
	RoleCanvas:               "canvas",
	RoleCheckBox:             "check box",
	RoleCheckMenuItem:        "check menu item",
	RoleColorChooser:         "color chooser",
	RoleColumnHeader:         "column header",
	RoleComboBox:             "combo box",
	RoleDateEditor:           "date editor",
	RoleDesktopIcon:          "desktop icon",
	RoleDesktopFrame:         "desktop frame",
	RoleDial:                 "dial",
	RoleDialog:               "dialog",
	RoleDirectoryPane:        "directory pane",
	RoleDrawingArea:          "drawing area",
	RoleFileChooser:          "file chooser",
	RoleFiller:               "filler",
	RoleFocusTraversable:     "focus traversable",
	RoleFontChooser:          "font chooser",
	RoleFrame:                "frame",
	RoleGlassPane:            "glass pane",
	RoleHtmlContainer:        "html container",
	RoleIcon:                 "icon",
	RoleImage:                "image",
	RoleInternalFrame:        "internal frame",
	RoleLabel:                "label",
	RoleLayeredPane:          "layered pane",
	RoleList:                 "list",
	RoleListItem:             "list item",
	RoleMenu:                 "menu",
	RoleMenuBar:              "menu bar",
	RoleMenuItem:             "menu item",
	RoleOptionPane:           "option pane",
	RolePageTab:              "page tab",
	RolePageTabList:          "page tab list",
	RolePanel:                "panel",
	RolePasswordText:         "password text",
	RolePopupMenu:            "popup menu",
	RoleProgressBar:          "progress bar",
	RoleButton:               "push button",
	RoleRadioButton:          "radio button",
	RoleRadioMenuItem:        "radio menu item",
	RoleRootPane:             "root pane",
	RoleRowHeader:            "row header",
	RoleScrollBar:            "scroll bar",
	RoleScrollPane:           "scroll pane",
	RoleSeparator:            "separator",
	RoleSlider:               "slider",
	RoleSpinButton:           "spin button",
	RoleSplitPane:            "split pane",
	RoleStatusBar:            "status bar",
	RoleTable:                "table",
	RoleTableCell:            "table cell",
	RoleTableColumnHeader:    "table column header",
	RoleTableRowHeader:       "table row header",
	RoleTearoffMenuItem:      "tearoff menu item",
	RoleTerminal:             "terminal",
	RoleText:                 "text",
	RoleToggleButton:         "toggle button",
	RoleToolBar:              "tool bar",
	RoleToolTip:              "tool tip",
	RoleTree:                 "tree",
	RoleTreeTable:            "tree table",
	RoleUnknown:              "unknown",
	RoleViewport:             "viewport",
	RoleWindow:               "window",
	RoleExtended:             "extended",
	RoleHeader:               "header",
	RoleFooter:               "footer",
	RoleParagraph:            "paragraph",
	RoleRuler:                "ruler",
	RoleApplication:          "application",
	RoleAutocomplete:         "autocomplete",
	RoleEditbar:              "editbar",
	RoleEmbedded:             "embedded",
	RoleEntry:                "entry",
	RoleChart:                "chart",
	RoleCaption:              "caption",
	RoleDocumentFrame:        "document frame",
	RoleHeading:              "heading",
	RolePage:                 "page",
	RoleSection:              "section",
	RoleRedundantObject:      "redundant object",
	RoleForm:                 "form",
	RoleLink:                 "link",
	RoleInputMethodWindow:    "input method window",
	RoleTableRow:             "table row",
	RoleTreeItem:             "tree item",
	RoleDocumentSpreadsheet:  "document spreadsheet",
	RoleDocumentPresentation: "document presentation",
	RoleDocumentText:         "document text",
	RoleDocumentWeb:          "document web",
	RoleDocumentEmail:        "document email",
	RoleComment:              "comment",
	RoleListBox:              "list box",
	RoleGrouping:             "grouping",
	RoleImageMap:             "image map",
	RoleNotification:         "notification",
	RoleInfoBar:              "info bar",
	RoleLevelBar:             "level bar",
	RoleTitleBar:             "title bar",
	RoleBlockQuote:           "block quote",
	RoleAudio:                "audio",
	RoleVideo:                "video",
	RoleDefinition:           "definition",
	RoleArticle:              "article",
	RoleLandmark:             "landmark",
	RoleLog:                  "log",
	RoleMarquee:              "marquee",
	RoleMath:                 "math",
	RoleRating:               "rating",
	RoleTimer:                "timer",
	RoleStatic:               "static",
	RoleMathFraction:         "math fraction",
	RoleMathRoot:             "math root",
	RoleSubscript:            "subscript",
	RoleSuperscript:          "superscript",
	RoleDescriptionList:      "description list",
	RoleDescriptionTerm:      "description term",
	RoleDescriptionValue:     "description value",
	RoleFootnote:             "footnote",
	RoleContentDeletion:      "content deletion",
	RoleContentInsertion:     "content insertion",
	RoleMark:                 "mark",
	RoleSuggestion:           "suggestion",
	RolePushButtonMenu:       "push button menu",
	RoleSwitch:               "switch",
}

// State is one of an object's states (AtspiStateType).
type State uint32

// The states.
const (
	StateInvalid                State = 0
	StateActive                 State = 1
	StateArmed                  State = 2
	StateBusy                   State = 3
	StateChecked                State = 4
	StateCollapsed              State = 5
	StateDefunct                State = 6
	StateEditable               State = 7
	StateEnabled                State = 8
	StateExpandable             State = 9
	StateExpanded               State = 10
	StateFocusable              State = 11
	StateFocused                State = 12
	StateHasTooltip             State = 13
	StateHorizontal             State = 14
	StateIconified              State = 15
	StateModal                  State = 16
	StateMultiLine              State = 17
	StateMultiselectable        State = 18
	StateOpaque                 State = 19
	StatePressed                State = 20
	StateResizable              State = 21
	StateSelectable             State = 22
	StateSelected               State = 23
	StateSensitive              State = 24
	StateShowing                State = 25
	StateSingleLine             State = 26
	StateStale                  State = 27
	StateTransient              State = 28
	StateVertical               State = 29
	StateVisible                State = 30
	StateManagesDescendants     State = 31
	StateIndeterminate          State = 32
	StateRequired               State = 33
	StateTruncated              State = 34
	StateAnimated               State = 35
	StateInvalidEntry           State = 36
	StateSupportsAutocompletion State = 37
	StateSelectableText         State = 38
	StateIsDefault              State = 39
	StateVisited                State = 40
	StateCheckable              State = 41
	StateHasPopup               State = 42
	StateReadOnly               State = 43
	StateDefault                State = 39
)
