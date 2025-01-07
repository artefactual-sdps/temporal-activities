package premis

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/beevik/etree"
	"go.artefactual.dev/tools/fsutil"
)

const (
	EmptyXML = `<?xml version="1.0" encoding="UTF-8"?>
<premis:premis xmlns:premis="http://www.loc.gov/premis/v3" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.loc.gov/premis/v3 https://www.loc.gov/standards/premis/premis.xsd" version="3.0"></premis:premis>
`
	indentSpaces = 2
)

type ObjectEventIdentifier struct {
	IdType  string
	IdValue string
}

type Object struct {
	IdType           string
	IdValue          string
	OriginalName     string
	EventIdentifiers []ObjectEventIdentifier
}

type EventSummary struct {
	IdType        string
	IdValue       string
	DateTime      string
	Type          string
	Detail        string
	Outcome       string
	OutcomeDetail string
}

type Event struct {
	IdType       string
	IdValue      string
	Summary      EventSummary
	DateTime     string
	AgentIdType  string
	AgentIdValue string
}

type Agent struct {
	IdType  string
	IdValue string
	Name    string
	Type    string
}

func AgentDefault() Agent {
	return Agent{
		Type:    "software",
		Name:    "Enduro",
		IdType:  "url",
		IdValue: "https://github.com/artefactual-sdps/preprocessing-sfa",
	}
}

func NewDoc() (*etree.Document, error) {
	doc := newDoc()

	err := doc.ReadFromString(EmptyXML)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func ParseFile(filePath string) (*etree.Document, error) {
	doc := newDoc()

	err := doc.ReadFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("parse XML: %v", err)
	}

	return doc, nil
}

func ParseOrInitialize(filePath string) (*etree.Document, error) {
	var doc *etree.Document
	var err error

	exists, err := fsutil.Exists(filePath)
	if err != nil {
		return nil, err
	}

	if exists {
		doc, err = ParseFile(filePath)
	} else {
		doc, err = NewDoc()
	}
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func WriteIndentedToFile(doc *etree.Document, filePath string) error {
	doc.Indent(indentSpaces)
	return doc.WriteToFile(filePath)
}

func WriteIndentedToString(doc *etree.Document) (string, error) {
	doc.Indent(indentSpaces)
	return doc.WriteToString()
}

func AppendObjectXML(doc *etree.Document, object Object) error {
	el, err := getRoot(doc)
	if err != nil {
		return err
	}

	addObjectElementIfNeeded(el, object)

	return nil
}

func AppendEventXMLForEachObject(doc *etree.Document, eventSummary EventSummary, agent Agent) error {
	PREMISEl, err := getRoot(doc)
	if err != nil {
		return err
	}

	// Add events for each existing object.
	for _, objectEl := range PREMISEl.FindElements("//premis:object") {
		// Define PREMIS event.
		event := eventFromEventSummaryAndAgent(eventSummary, agent)

		// Add PREMIS event element and, if necessary, agent element.
		addEventElement(PREMISEl, event)

		// Link event to object
		LinkEventToObject(objectEl, event)
	}

	return nil
}

func AppendAgentXML(doc *etree.Document, agent Agent) error {
	el, err := getRoot(doc)
	if err != nil {
		return err
	}

	addAgentElementIfNeeded(el, agent)

	return nil
}

func LinkEventToObject(objectEl *etree.Element, eventFull Event) {
	linkEventIdEl := objectEl.CreateElement("premis:linkingEventIdentifier")

	linkEventIdTypeEl := linkEventIdEl.CreateElement("premis:linkingEventIdentifierType")
	linkEventIdTypeEl.CreateText(eventFull.Summary.IdType)

	linkEventIdValueEl := linkEventIdEl.CreateElement("premis:linkingEventIdentifierValue")
	linkEventIdValueEl.CreateText(eventFull.Summary.IdValue)
}

func FilesWithinDirectory(contentPath string) ([]string, error) {
	var subpaths []string

	err := filepath.WalkDir(contentPath, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		subpath, err := filepath.Rel(contentPath, p)
		if err != nil {
			return err
		}

		subpaths = append(subpaths, subpath)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return subpaths, nil
}

func newDoc() *etree.Document {
	doc := etree.NewDocument()
	doc.WriteSettings = etree.WriteSettings{CanonicalEndTags: true}

	return doc
}

func getRoot(doc *etree.Document) (*etree.Element, error) {
	// Get PREMIS root element.
	el := doc.FindElement("/premis:premis")
	if el == nil {
		return nil, errors.New("no root premis element found in document")
	}

	return el, nil
}

func eventFromEventSummaryAndAgent(eventSummary EventSummary, agent Agent) Event {
	return Event{
		Summary:      eventSummary,
		AgentIdType:  agent.IdType,
		AgentIdValue: agent.IdValue,
	}
}

func addObjectElementIfNeeded(PREMISEl *etree.Element, object Object) {
	if checkIfObjectElementExists(PREMISEl, object) {
		return
	}

	objectEl := PREMISEl.CreateElement("premis:object")
	objectEl.CreateAttr("xsi:type", "premis:file")

	// Add object identifier elements.
	objectIdEl := objectEl.CreateElement("premis:objectIdentifier")

	objectIdentifierTypeEl := objectIdEl.CreateElement("premis:objectIdentifierType")
	objectIdentifierTypeEl.CreateText(object.IdType)

	objectIdentifierValueEl := objectIdEl.CreateElement("premis:objectIdentifierValue")
	objectIdentifierValueEl.CreateText(object.IdValue)

	// Add object characteristics element.
	objectCharEl := objectEl.CreateElement("premis:objectCharacteristics")

	formatEl := objectCharEl.CreateElement("premis:format")

	formatDesEl := formatEl.CreateElement("premis:formatDesignation")

	formatDesEl.CreateElement("premis:formatName")

	// Add original name element.
	originalNameEl := objectEl.CreateElement("premis:originalName")
	originalNameEl.CreateText(object.OriginalName)
}

func addEventElement(PREMISEl *etree.Element, event Event) {
	eventEl := PREMISEl.CreateElement("premis:event")

	// Add event identifier elements.
	eventIdEl := eventEl.CreateElement("premis:eventIdentifier")

	eventIdentifierTypeEl := eventIdEl.CreateElement("premis:eventIdentifierType")
	eventIdentifierTypeEl.CreateText(event.Summary.IdType)

	eventIdentifierValueEl := eventIdEl.CreateElement("premis:eventIdentifierValue")
	eventIdentifierValueEl.CreateText(event.Summary.IdValue)

	// Add event type and datetime elements.
	eventTypeEl := eventEl.CreateElement("premis:eventType")
	eventTypeEl.CreateText(event.Summary.Type)

	eventDateEl := eventEl.CreateElement("premis:eventDateTime")
	eventDateEl.CreateText(event.Summary.DateTime)

	// Add event detail elements.
	eventDetailInfoEl := eventEl.CreateElement("premis:eventDetailInformation")
	eventDetailEl := eventDetailInfoEl.CreateElement("premis:eventDetail")
	eventDetailEl.CreateText(event.Summary.Detail)

	// Add event outcome elements.
	outcomeInfoEl := eventEl.CreateElement("premis:eventOutcomeInformation")
	outcomeEl := outcomeInfoEl.CreateElement("premis:eventOutcome")
	outcomeEl.CreateText(event.Summary.Outcome)

	if event.Summary.OutcomeDetail != "" {
		outcomeDetailEl := outcomeInfoEl.CreateElement("premis:eventOutcomeDetail")
		outcomeDetailNoteEl := outcomeDetailEl.CreateElement("premis:eventOutcomeDetailNote")
		outcomeDetailNoteEl.CreateText(event.Summary.OutcomeDetail)
	}

	addEventAgentIdentifierElement(eventEl, event)
}

func addEventAgentIdentifierElement(eventEl *etree.Element, event Event) {
	linkAgentIdentifierEl := eventEl.CreateElement("premis:linkingAgentIdentifier")

	// Add linking agent identifier type element.
	linkAgentIdentifierTypeEl := linkAgentIdentifierEl.CreateElement("premis:linkingAgentIdentifierType")
	linkAgentIdentifierTypeEl.CreateText(event.AgentIdType)
	linkAgentIdentifierTypeEl.CreateAttr("valueURI", "http://id.loc.gov/vocabulary/identifiers/local")

	// Add linking agent identifier value element.
	linkAgentIdentifierValueEl := linkAgentIdentifierEl.CreateElement("premis:linkingAgentIdentifierValue")
	linkAgentIdentifierValueEl.CreateText(event.AgentIdValue)
}

func addAgentElementIfNeeded(PREMISEl *etree.Element, agent Agent) {
	if checkIfAgentElementExists(PREMISEl, agent) {
		return
	}

	// Add agent element.
	agentEl := PREMISEl.CreateElement("premis:agent")

	// Add agent identifier elements.
	agentIdentifierEl := agentEl.CreateElement("premis:agentIdentifier")

	agentIdentifierTypeEl := agentIdentifierEl.CreateElement("premis:agentIdentifierType")
	agentIdentifierTypeEl.CreateText(agent.IdType)
	agentIdentifierTypeEl.CreateAttr("valueURI", "http://id.loc.gov/vocabulary/identifiers/local")

	agentIdentifierValueEl := agentIdentifierEl.CreateElement("premis:agentIdentifierValue")
	agentIdentifierValueEl.CreateText(agent.IdValue)

	// Add agent name and type.
	agentNameEl := agentEl.CreateElement("premis:agentName")
	agentNameEl.CreateText(agent.Name)

	agentTypeEl := agentEl.CreateElement("premis:agentType")
	agentTypeEl.CreateText(agent.Type)
}

func checkForDuplicateElementData(PREMISEl *etree.Element, elementTag string, paths map[string]string) bool {
	// Cycle through agent elements so we can compare child data.
	elements := PREMISEl.FindElements(fmt.Sprintf("//premis:%s", elementTag))

	for _, element := range elements {
		foundDifference := false

		// Check child fields to see if they all match.
		for path, value := range paths {
			childEl := element.FindElement(fmt.Sprintf(".//%s[text()='%s']", path, value))

			if childEl == nil {
				foundDifference = true

				break
			}
		}

		if !foundDifference {
			// Found match.
			return true
		}
	}

	// Found no match.
	return false
}

func checkIfObjectElementExists(PREMISEl *etree.Element, object Object) bool {
	// Define xpath paths to child elements and values to match on.
	paths := map[string]string{
		"premis:originalName": object.OriginalName,
	}

	return checkForDuplicateElementData(PREMISEl, "object", paths)
}

func checkIfAgentElementExists(PREMISEl *etree.Element, agent Agent) bool {
	// Define xpath paths to child elements and values to match on.
	paths := map[string]string{
		"premis:agentType": agent.Type,
		"premis:agentName": agent.Name,
		"premis:agentIdentifier/premis:agentIdentifierType":  agent.IdType,
		"premis:agentIdentifier/premis:agentIdentifierValue": agent.IdValue,
	}

	return checkForDuplicateElementData(PREMISEl, "agent", paths)
}
