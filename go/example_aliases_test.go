// SPDX-License-Identifier: EUPL-1.2

package api

type InputValidator = toolInputValidator
type ResponseRecorder = toolResponseRecorder
type MetaRecorder = responseMetaRecorder

type Number = jsonNumber
type RawMessage = jsonRawMessage
type Value = jsonValue

type StopList = chatStopList
type ResolutionError = modelResolutionError
type CompletionsHandler = chatCompletionsHandler
type CompletionRequestError = chatCompletionRequestError

type URLError = blockedURLError
