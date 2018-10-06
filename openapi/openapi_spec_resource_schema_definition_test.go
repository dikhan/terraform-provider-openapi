package openapi

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetStatusId(t *testing.T) {
	Convey("Given a SpecSchemaDefinition that has an status property that is not an object", t, func() {
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:     statusDefaultPropertyName,
					Type:     typeString,
					ReadOnly: true,
				},
			},
		}

		Convey("When getStatusIdentifier method is called", func() {
			statuses, err := s.getStatusIdentifier()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the statuses returned should not be empty'", func() {
				So(statuses, ShouldNotBeEmpty)
			})
			Convey("Then the value returned should contain the name of the property 'statuses'", func() {
				So(statuses[0], ShouldEqual, statusDefaultPropertyName)
			})
		})
	})

	Convey("Given a swagger schema definition that DOES NOT have an 'status' property but has a property configured with IsStatusIdentifier set to TRUE", t, func() {
		expectedStatusProperty := "some-other-property-holding-status"
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:               expectedStatusProperty,
					Type:               typeString,
					ReadOnly:           true,
					IsStatusIdentifier: true,
				},
			},
		}
		Convey("When getStatusIdentifier method is called", func() {
			status, err := s.getStatusIdentifier()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the status returned should not be empty'", func() {
				So(status, ShouldNotBeEmpty)
			})
			Convey("Then the value returned should contain the name of the property 'some-other-property-holding-status'", func() {
				So(status[0], ShouldEqual, expectedStatusProperty)
			})
		})
	})

	Convey("Given a swagger schema definition that HAS BOTH an 'status' property AND ALSO a property configured with 'x-terraform-field-status' set to true", t, func() {
		expectedStatusProperty := "some-other-property-holding-status"
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:     statusDefaultPropertyName,
					Type:     typeString,
					ReadOnly: true,
				},
				&specSchemaDefinitionProperty{
					Name:               expectedStatusProperty,
					Type:               typeString,
					ReadOnly:           true,
					IsStatusIdentifier: true,
				},
			},
		}
		Convey("When getStatusIdentifier method is called", func() {
			status, err := s.getStatusIdentifier()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the status returned should not be empty'", func() {
				So(status, ShouldNotBeEmpty)
			})
			Convey("Then the value returned should be 'some-other-property-holding-status' as it takes preference over the default 'status' property", func() {
				So(status[0], ShouldEqual, expectedStatusProperty)
			})
		})
	})

	Convey("Given a swagger schema definition that HAS an status field which is an object containing a status field", t, func() {
		expectedStatusProperty := "actualStatus"
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:     "id",
					Type:     typeString,
					ReadOnly: true,
				},
				&specSchemaDefinitionProperty{
					Name:     statusDefaultPropertyName,
					Type:     typeObject,
					ReadOnly: true,
					SpecSchemaDefinition: &specSchemaDefinition{
						Properties: specSchemaDefinitionProperties{
							&specSchemaDefinitionProperty{
								Name:               expectedStatusProperty,
								Type:               typeString,
								ReadOnly:           true,
								IsStatusIdentifier: true,
							},
						},
					},
				},
			},
		}
		Convey("When getStatusIdentifier method is called", func() {
			status, err := s.getStatusIdentifier()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the status returned should not be empty'", func() {
				So(status, ShouldNotBeEmpty)
			})
			Convey("Then the status array returned should contain the different the trace of property names (hierarchy) until the last one which is the one used as status, in this case 'actualStatus' on the last index", func() {
				So(status[0], ShouldEqual, "status")
				So(status[1], ShouldEqual, expectedStatusProperty)
			})
		})
	})

	Convey("Given a swagger schema definition that DOES NOT have an 'status' property but has a property configured with 'x-terraform-field-status' set to FALSE", t, func() {
		expectedStatusProperty := "some-other-property-holding-status"
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:               expectedStatusProperty,
					Type:               typeString,
					ReadOnly:           true,
					IsStatusIdentifier: false,
				},
			},
		}
		Convey("When getStatusIdentifier method is called", func() {
			_, err := s.getStatusIdentifier()
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a swagger schema definition that NEITHER HAS an 'status' property NOR a property configured with 'x-terraform-field-status' set to true", t, func() {
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:               "prop-that-is-not-status",
					Type:               typeString,
					ReadOnly:           true,
					IsStatusIdentifier: false,
				},
				&specSchemaDefinitionProperty{
					Name:               "prop-that-is-not-status-and-does-not-have-status-metadata-either",
					Type:               typeString,
					ReadOnly:           true,
					IsStatusIdentifier: false,
				},
			},
		}
		Convey("When getStatusIdentifier method is called", func() {
			_, err := s.getStatusIdentifier()
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a swagger schema definition with a status property that is not readonly", t, func() {
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:     statusDefaultPropertyName,
					Type:     typeString,
					ReadOnly: false,
				},
			},
		}
		Convey("When getStatusIdentifier method is called", func() {
			_, err := s.getStatusIdentifier()
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

}

func TestGetStatusIdentifierFor(t *testing.T) {
	Convey("Given a swagger schema definition with a property configured with 'x-terraform-field-status' set to true and it is not readonly", t, func() {
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:               statusDefaultPropertyName,
					Type:               typeString,
					ReadOnly:           false,
					IsStatusIdentifier: true,
				},
			},
		}
		Convey("When getStatusIdentifierFor method is called with a schema definition and forceReadOnly check is disabled (this happens when the method is called recursively)", func() {
			status, err := s.getStatusIdentifierFor(s, true, false)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the status array returned should contain the status property even though it's not read only...readonly checks are only perform on root level properties", func() {
				So(status[0], ShouldEqual, "status")
			})
		})
	})
}
