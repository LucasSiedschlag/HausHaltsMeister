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
    title: "Meios de Pagamento",
    href: "/meios-de-pagamento",
    icon: "CreditCard",
  },
  {
    title: "Picuinhas",
    href: "/picuinhas/pessoas",
    icon: "Users",
    children: [
      {
        title: "Pessoas",
        href: "/picuinhas/pessoas",
      },
      {
        title: "Lançamentos",
        href: "/picuinhas/lancamentos",
      },
    ],
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
