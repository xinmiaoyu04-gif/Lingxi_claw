package router_test

import (
	"testing"

	"lingxi-claw/internal/model"
	"lingxi-claw/internal/router"
)

func TestRouteFile(t *testing.T) {
	textPDF := []byte("%PDF-1.7\n1 0 obj<</Font<</F1 2 0 R>>>>\nstream\nBT (hello) Tj ET\nendstream\n")
	scanPDF := []byte("%PDF-1.7\n1 0 obj<</XObject<</Im0 2 0 R>>>>\nstream\n\xff\xd8\xff\xe0binary\nendstream\n")

	cases := []struct {
		name      string
		file      string
		content   []byte
		wantRoute string
		supported bool
	}{
		{"docx", "2024期末.docx", []byte("PK\x03\x04"), model.FileRouteDocxParser, true},
		{"text pdf", "2023期末.pdf", textPDF, model.FileRouteTextParser, true},
		{"scanned pdf", "扫描版.pdf", scanPDF, model.FileRouteOCR, true},
		{"jpg", "homework.jpg", []byte("\xff\xd8\xff"), model.FileRouteOCR, true},
		{"png uppercase ext", "HOMEWORK.PNG", []byte("\x89PNG"), model.FileRouteOCR, true},
		{"unsupported txt", "notes.txt", []byte("plain"), "", false},
		{"no extension", "scan", []byte("x"), "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := router.RouteFile(tc.file, tc.content)
			if got.Supported != tc.supported {
				t.Fatalf("Supported = %v, want %v", got.Supported, tc.supported)
			}
			if got.FileRoute != tc.wantRoute {
				t.Errorf("FileRoute = %q, want %q", got.FileRoute, tc.wantRoute)
			}
		})
	}
}

func TestHasTextLayer(t *testing.T) {
	if router.HasTextLayer(nil) {
		t.Error("empty content reported a text layer")
	}
	if !router.HasTextLayer([]byte("stream\nBT /F1 12 Tf (x) Tj ET\nendstream")) {
		t.Error("content with Tj operators reported no text layer")
	}
	if router.HasTextLayer([]byte("stream\n\xff\xd8\xff\xe0 raw jpeg\nendstream")) {
		t.Error("image-only content reported a text layer")
	}
}

func TestRouteText(t *testing.T) {
	cases := []struct {
		name           string
		message        string
		wantComplexity string
		wantModel      string
	}{
		{"short factual", "导数是啥", router.ComplexityLow, router.ModelLightweight},
		{"calculation ask", "帮我计算这个二重积分", router.ComplexityMedium, router.ModelStandard},
		{"proof ask", "请证明这个函数连续", router.ComplexityHigh, router.ModelStandard},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := router.RouteText(tc.message)
			if got.Complexity != tc.wantComplexity {
				t.Errorf("Complexity = %q, want %q", got.Complexity, tc.wantComplexity)
			}
			if got.ModelRoute != tc.wantModel {
				t.Errorf("ModelRoute = %q, want %q", got.ModelRoute, tc.wantModel)
			}
		})
	}
}

func TestRouteImageUsesMultimodalTier(t *testing.T) {
	got := router.RouteImage()
	if got.ModelRoute != router.ModelMultimodal {
		t.Errorf("ModelRoute = %q, want %q", got.ModelRoute, router.ModelMultimodal)
	}
	if got.Tool != "vision_model" {
		t.Errorf("Tool = %q, want vision_model", got.Tool)
	}
}

func TestClassifyQuestion(t *testing.T) {
	cases := map[string]string{
		"证明：连续函数有界":       router.QuestionTypeProof,
		"计算二重积分 ∬ x dxdy": router.QuestionTypeCalculation,
		"下列说法正确的是":        router.QuestionTypeChoice,
		"填空：极限值为 ______":  router.QuestionTypeFill,
		"什么是拉格朗日中值定理":     router.QuestionTypeConcept,
	}
	for content, want := range cases {
		if got := router.ClassifyQuestion(content); got != want {
			t.Errorf("ClassifyQuestion(%q) = %q, want %q", content, got, want)
		}
	}
}

func TestQuestionTypeNameIsChinese(t *testing.T) {
	if got := router.QuestionTypeName(router.QuestionTypeCalculation); got != "计算题" {
		t.Errorf("calculation label = %q, want 计算题", got)
	}
	if got := router.QuestionTypeName("unknown_slug"); got != "其它" {
		t.Errorf("unknown slug label = %q, want 其它", got)
	}
}
