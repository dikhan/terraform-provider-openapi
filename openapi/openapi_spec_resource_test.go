package openapi

//
//func TestIsResourcePollingEnabled(t *testing.T) {
//	Convey("Given a resourceInfo", t, func() {
//		r := SpecV2Resource{}
//		Convey("When isResourcePollingEnabled method is called with a list of responses where one of the reponses matches the response status received and has the 'x-terraform-resource-poll-enabled' extension set to true", func() {
//			extensions := spec.Extensions{}
//			extensions.Add(extTfResourcePollEnabled, true)
//			responses := &spec.Responses{
//				ResponsesProps: spec.ResponsesProps{
//					StatusCodeResponses: map[int]spec.Response{
//						http.StatusAccepted: {
//							VendorExtensible: spec.VendorExtensible{
//								Extensions: extensions,
//							},
//						},
//					},
//				},
//			}
//			isResourcePollingEnabled, _ := r.isResourcePollingEnabled(responses, http.StatusAccepted)
//			Convey("Then the bool returned should be true", func() {
//				So(isResourcePollingEnabled, ShouldBeTrue)
//			})
//		})
//		Convey("When isResourcePollingEnabled method is called with a list of responses where one of the reponses matches the response status received and has the 'x-terraform-resource-poll-enabled' extension set to false", func() {
//			extensions := spec.Extensions{}
//			extensions.Add(extTfResourcePollEnabled, false)
//			responses := &spec.Responses{
//				ResponsesProps: spec.ResponsesProps{
//					StatusCodeResponses: map[int]spec.Response{
//						http.StatusAccepted: {
//							VendorExtensible: spec.VendorExtensible{
//								Extensions: extensions,
//							},
//						},
//					},
//				},
//			}
//			isResourcePollingEnabled, _ := r.isResourcePollingEnabled(responses, http.StatusAccepted)
//			Convey("Then the bool returned should be false", func() {
//				So(isResourcePollingEnabled, ShouldBeFalse)
//			})
//		})
//		Convey("When isResourcePollingEnabled method is called with list of responses where non of the codes match the given response http code", func() {
//			responses := &spec.Responses{
//				ResponsesProps: spec.ResponsesProps{
//					StatusCodeResponses: map[int]spec.Response{
//						http.StatusOK: {},
//					},
//				},
//			}
//			isResourcePollingEnabled, _ := r.isResourcePollingEnabled(responses, http.StatusAccepted)
//			Convey("Then bool returned should be false", func() {
//				So(isResourcePollingEnabled, ShouldBeFalse)
//			})
//		})
//	})
//}
