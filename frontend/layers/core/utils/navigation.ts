export interface NavigationItem {
  title: string
  href: string
  icon?: string
  children?: NavigationItem[]
}

export const mainNavigation: NavigationItem[] = [
  {
    title: "Dashboard",
    href: "/",
    icon: "LayoutDashboard",
  },
  {
    title: "Categorias",
    href: "/categorias",
    icon: "Tags",
  },
  {
    title: "Orçamento",
    href: "/orcamento",
    icon: "Wallet",
  },
  {
    title: "Transactions",
    href: "/transactions",
    icon: "DollarSign",
  },
  {
    title: "Reports",
    href: "/reports",
    icon: "FileBarChart",
  }
]
