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
    title: "Lançamentos",
    href: "/lancamentos/entradas",
    icon: "ArrowDownUp",
    children: [
      {
        title: "Entradas",
        href: "/lancamentos/entradas",
      },
      {
        title: "Fixos",
        href: "/lancamentos/fixos",
      },
      {
        title: "Variáveis",
        href: "/lancamentos/variaveis",
      },
      {
        title: "Picuinhas",
        href: "/lancamentos/picuinhas",
      },
    ],
  },
  {
    title: "Meios de Pagamento",
    href: "/meios-de-pagamento",
    icon: "CreditCard",
  },
  {
    title: "Parcelamentos",
    href: "/parcelamentos/nova-compra",
    icon: "Receipt",
    children: [
      {
        title: "Nova compra parcelada",
        href: "/parcelamentos/nova-compra",
      },
      {
        title: "Fatura de cartão",
        href: "/parcelamentos/fatura",
      },
    ],
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
    title: "Reports",
    href: "/reports",
    icon: "FileBarChart",
  }
]
