package model

type NodeType string

const (
	NodeTypeProgram            NodeType = "program"
	NodeTypeParagraph          NodeType = "paragraph"
	NodeTypeGoto               NodeType = "goto"
	NodeTypePerform            NodeType = "perform"
	NodeTypeSection            NodeType = "section"
	NodeTypeStop               NodeType = "stop"
	NodeTypeExit               NodeType = "exit"
	NodeTypeExitSection        NodeType = "exitsection"
	NodeTypeExitParagraph      NodeType = "exitparagraph"
	NodeTypeGoback             NodeType = "goback"
	NodeTypeIf                 NodeType = "if"
	NodeTypeElse               NodeType = "else"
	NodeTypeEndif              NodeType = "endif"
	NodeTypeEvaluate           NodeType = "evaluate"
	NodeTypeWhen               NodeType = "when"
	NodeTypeWhenOther          NodeType = "whenother"
	NodeTypeEndEvaluate        NodeType = "endevaluate"
	NodeTypeInlinePerform      NodeType = "inlineperform"
	NodeTypeEndInlinePerform   NodeType = "endinlineperform"
	NodeTypeAtEnd              NodeType = "atEnd"
	NodeTypeAtEndExit          NodeType = "atEndExit"
	NodeTypeAlter              NodeType = "alter"
	NodeTypeOutput             NodeType = "output"
	NodeTypeInput              NodeType = "input"
	NodeTypeSort               NodeType = "sort"
	NodeTypeEndSort            NodeType = "endsort"
	NodeTypeMerge              NodeType = "merge"
	NodeTypeEndMerge           NodeType = "endmerge"
	NodeTypeXmlParse           NodeType = "xmlparse"
	NodeTypeEndXml             NodeType = "endxml"
	NodeTypeOnException        NodeType = "onexception"
	NodeTypeOnNotException     NodeType = "onnotexception"
	NodeTypeEndOn              NodeType = "endon"
	NodeTypeStatement          NodeType = "statement"
	NodeTypeExecSql            NodeType = "execsql"
	NodeTypeExecWhenever       NodeType = "execwhenever"
	NodeTypeExecCics           NodeType = "execcics"
	NodeTypeExecCicsReturn     NodeType = "execcicsreturn"
	NodeTypeExecCicsHandle     NodeType = "execcicshandle"
	NodeTypeExecCicsAbend      NodeType = "execcicsabend"
	NodeTypeEndExec            NodeType = "endexec"
	NodeTypeUse                NodeType = "use"
	NodeTypeUseForDebugging    NodeType = "usefordebugging"
	NodeTypeExitPerform        NodeType = "exitperform"
	NodeTypeCall               NodeType = "call"
)

type Position struct {
	Line      int32 `json:"line"`
	Character int32 `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Location struct {
	Uri string `json:"uri"`
	Range
}

type ProcedureName struct {
	Name      string  `json:"name"`
	InSection *string `json:"inSection,omitempty"`
}

type CFASTNode struct {
	ID        int32        `json:"id,omitempty"`
	Children  []*CFASTNode `json:"children,omitempty"`
	Type      NodeType     `json:"type"`
	Location  Location     `json:"location"`
	Processed bool         `json:"processed,omitempty"`
	Snippet   string       `json:"snippet,omitempty"`

	// Fields for specific types
	Name              string         `json:"name,omitempty"`
	TargetName        interface{}    `json:"targetName,omitempty"` // string or []string (for Goto)
	TargetSectionName *string        `json:"targetSectionName,omitempty"`
	ThruName          *string        `json:"thruName,omitempty"`
	ThruSectionName   *string        `json:"thruSectionName,omitempty"`
	PerformUntilType  *string        `json:"performUntilType,omitempty"`
	Cycle             *bool          `json:"cycle,omitempty"`
	InsideInlinePerform *bool        `json:"insideInlinePerform,omitempty"`
	From              *ProcedureName `json:"from,omitempty"`
	To                *ProcedureName `json:"to,omitempty"`
	Target            *ProcedureName `json:"target,omitempty"`
	Thru              *ProcedureName `json:"thru,omitempty"`
	HandleType        *string        `json:"handleType,omitempty"`
	Value             *string        `json:"value,omitempty"`
	Cancel            *bool          `json:"cancel,omitempty"`
	WheneverCondition *string        `json:"wheneverCondition,omitempty"`
	WheneverType      *string        `json:"wheneverType,omitempty"`
}
