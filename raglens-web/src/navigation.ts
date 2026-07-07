export type NavigationItem = {
  label: string;
  to: "/" | "/documents" | "/ask" | "/traces";
  description: string;
};

export const navigationItems: NavigationItem[] = [
  { label: "Overview", to: "/", description: "运行状态和当前工作流摘要" },
  { label: "Documents", to: "/documents", description: "上传、查看 chunks、触发 index" },
  { label: "Ask", to: "/ask", description: "检索调试和 RAG Ask" },
  { label: "Traces", to: "/traces", description: "查看 RAG 调用链路" },
];

export function resolveNavigationItem(pathname: string) {
  return navigationItems.find((item) => item.to === pathname) ?? navigationItems[0];
}
