package repository

import (
	_ "embed"
	"fmt"
	"sort"
	"strings"
	"text/template"

	pgs "github.com/lyft/protoc-gen-star"
	pgsgo "github.com/lyft/protoc-gen-star/lang/go"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	protos "github.com/SpeedyCoder/protobolt/proto/v1"
)

//go:embed template.tmpl
var templateAsset string

const module = "protobolt"

// IdentifierModule validates & generates code for accessing an identifier on a message
type IdentifierModule struct {
	*pgs.ModuleBase
	ctx       pgsgo.Context
	entityTpl *template.Template
	importTpl *template.Template
}

type entity struct {
	Name     string
	PKFields []pgs.Field
}

func (e entity) RepositoryName() string {
	return e.Name + "Repository"
}

// NewRepositoryModule creates a module for PG*
func NewRepositoryModule() *IdentifierModule {
	return &IdentifierModule{
		ModuleBase: &pgs.ModuleBase{},
	}
}

// Name identifies this module
func (m *IdentifierModule) Name() string {
	return module
}

// InitContext sets up module for use
func (m *IdentifierModule) InitContext(c pgs.BuildContext) {
	m.ModuleBase.InitContext(c)
	m.ctx = pgsgo.InitContext(c.Parameters())

	tpl := template.New(module).Funcs(map[string]interface{}{
		"package":         m.ctx.PackageName,
		"pkFieldToString": pkFieldToString,
	})
	m.entityTpl = template.Must(tpl.Parse(templateAsset))
}

// Execute runs the generator
func (m *IdentifierModule) Execute(targets map[string]pgs.File, pkgs map[string]pgs.Package) []pgs.Artifact {
	for _, t := range targets {
		m.Debugf("generating for target: %s", t.Name())

		entities, err := m.generateEntities(t)
		if err != nil {
			m.AddError(err.Error())
			break
		}

		m.render(t, entities)
	}

	return m.Artifacts()
}

func (m *IdentifierModule) generateEntities(f pgs.File) ([]entity, error) {
	if len(f.Messages()) == 0 {
		m.Debugf("zero messages, skipping: %s", f.Name())
		return nil, nil
	}
	entities := make([]entity, 0, len(f.Messages()))

	for _, msg := range f.Messages() {

		e := entity{Name: msg.Name().String()}

		for _, field := range msg.Fields() {
			fdesc := field.Descriptor()
			if fdesc == nil || fdesc.GetOptions() == nil {
				continue
			}

			if !proto.HasExtension(fdesc.GetOptions(), protos.E_PrimaryKey) {
				continue
			}

			if proto.GetExtension(fdesc.GetOptions(), protos.E_PrimaryKey).(bool) {
				if !validFieldType(field) {
					return entities, errors.Errorf("unable to handle identifier field type: %s", field.Type().ProtoType())
				}

				e.PKFields = append(e.PKFields, field)
			}
		}
		if len(e.PKFields) != 0 {
			sort.Slice(e.PKFields, func(i, j int) bool {
				return *e.PKFields[i].Descriptor().Number < *e.PKFields[j].Descriptor().Number
			})
			entities = append(entities, e)
		}
		strings.Join(
			[]string{
				"",
			},
			"",
		)
	}

	return entities, nil
}

func (m *IdentifierModule) render(f pgs.File, entities []entity) {
	if len(entities) == 0 {
		return
	}

	name := m.ctx.OutputPath(f).SetExt(".protobolt.go")
	m.AddGeneratorTemplateFile(name.String(), m.entityTpl, map[string]interface{}{
		"File":     f,
		"Entities": entities,
	})
}

func validFieldType(field pgs.Field) bool {
	switch field.Type().ProtoType() {
	case pgs.StringT:
		return true
	case pgs.EnumT:
		return true
	case pgs.BoolT:
		return true
	case pgs.Int32T:
		return true
	case pgs.Int64T:
		return true
	case pgs.UInt32T:
		return true
	case pgs.UInt64T:
		return true
	}
	return false
}

func pkFieldToString(field pgs.Field) string {
	switch field.Type().ProtoType() {
	case pgs.StringT:
		return "m." + field.Name().UpperCamelCase().String()
	case pgs.EnumT:
		return "strconv.FormatInt(int64(m." + field.Name().UpperCamelCase().String() + "), 10)"
	case pgs.BoolT:
		return "strconv.FormatBool(m." + field.Name().UpperCamelCase().String() + ")"
	case pgs.Int32T:
		return "strconv.FormatInt(int64(m." + field.Name().UpperCamelCase().String() + "), 10)"
	case pgs.Int64T:
		return "strconv.FormatInt(m." + field.Name().UpperCamelCase().String() + ", 10)"
	case pgs.UInt32T:
		return "strconv.FormatUint(int64(m." + field.Name().UpperCamelCase().String() + "), 10)"
	case pgs.UInt64T:
		return "strconv.FormatUint(m." + field.Name().UpperCamelCase().String() + ", 10)"
	}
	panic(fmt.Sprintf("unexpected field type: %s", field.Type().ProtoType().String()))
}
