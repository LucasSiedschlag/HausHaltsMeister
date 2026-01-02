<script setup lang="ts">
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail
} from '@shared/components/ui/sidebar'
import { Calendar, CreditCard, LayoutDashboard, LineChart, Send, Settings, Tag, Wallet } from 'lucide-vue-next'
import NavMain from './NavMain.vue'
import NavProjects from './NavProjects.vue'
import NavSecondary from './NavSecondary.vue'
import NavUser from './NavUser.vue'

const navMain = [
  { title: 'Visão geral', url: '/', icon: LayoutDashboard },
  { title: 'Transações', url: '/journal', icon: Wallet },
  { title: 'Contas', url: '/accounts', icon: Calendar },
  { title: 'Cartões', url: '/credit-cards', icon: CreditCard },
  { title: 'Orçamentos', url: '/budgets', icon: LineChart },
  { title: 'Investimentos', url: '/investments', icon: Wallet }
]

const navProjects = [
  { title: 'Ledger principal', url: '/ledgers/current', icon: Tag },
  { title: 'Relatórios', url: '/reports', icon: LineChart }
]

const navSecondary = [
  { title: 'Categorias', url: '/categories', icon: Tag },
  { title: 'Configurações', url: '/settings', icon: Settings }
]
</script>

<template>
  <Sidebar collapsible="icon">
    <SidebarHeader>
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton size="lg" as-child>
            <NuxtLink to="/">
              <div class="flex h-9 w-9 items-center justify-center rounded-xl bg-primary text-primary-foreground text-xs font-semibold">
                HH
              </div>
              <div class="grid text-left text-sm leading-tight">
                <span class="font-semibold">HausHaltsMeister</span>
                <span class="text-xs text-muted-foreground">Ledger Finance</span>
              </div>
            </NuxtLink>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarHeader>

    <SidebarContent>
      <NavMain :items="navMain" />
      <SidebarGroup>
        <SidebarGroupLabel>Projetos</SidebarGroupLabel>
        <SidebarGroupContent>
          <NavProjects :items="navProjects" />
        </SidebarGroupContent>
      </SidebarGroup>
      <NavSecondary :items="navSecondary" />
    </SidebarContent>

    <SidebarFooter class="space-y-2">
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton as-child>
            <NuxtLink to="/feedback">
              <Send />
              <span>Feedback</span>
            </NuxtLink>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
      <NavUser />
    </SidebarFooter>

    <SidebarRail />
  </Sidebar>
</template>
