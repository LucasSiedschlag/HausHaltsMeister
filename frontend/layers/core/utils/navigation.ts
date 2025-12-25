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
