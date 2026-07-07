import { useEffect, useRef } from "react";
import { json } from "@codemirror/lang-json";
import { basicSetup, EditorView } from "codemirror";

type CodeViewerProps = {
  value: string;
  language?: "json" | "text";
  className?: string;
};

const viewerTheme = EditorView.theme({
  "&": {
    backgroundColor: "#0d0e0c",
    color: "#d7d5cb",
    fontSize: "12px",
    minHeight: "100%",
  },
  ".cm-content": {
    fontFamily: '"Cascadia Code", "JetBrains Mono", Consolas, monospace',
    padding: "12px",
  },
  ".cm-gutters": {
    backgroundColor: "#11130f",
    color: "#716d62",
    borderRight: "1px solid rgba(242, 240, 233, 0.1)",
  },
  ".cm-activeLine, .cm-activeLineGutter": {
    backgroundColor: "rgba(200, 255, 114, 0.06)",
  },
  ".cm-scroller": {
    overflow: "auto",
  },
});

export function CodeViewer({ value, language = "text", className = "" }: CodeViewerProps) {
  const hostRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    const host = hostRef.current;
    if (!host) return;

    const view = new EditorView({
      doc: value,
      extensions: [
        basicSetup,
        viewerTheme,
        EditorView.editable.of(false),
        EditorView.lineWrapping,
        ...(language === "json" ? [json()] : []),
      ],
      parent: host,
    });

    return () => view.destroy();
  }, [language, value]);

  return <div className={`code-viewer ${className}`.trim()} ref={hostRef} />;
}
