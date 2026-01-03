<script setup lang="ts">
import { computed } from 'vue'
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
import { useRoute } from '#imports'
import NavMain from './NavMain.vue'
import NavProjects from './NavProjects.vue'
import NavSecondary from './NavSecondary.vue'
import NavUser from './NavUser.vue'
import { useLedger } from '@shared/composables/useLedger'

const { t } = useI18n()
const { currentLedger, currentLedgerId, ledgers, fetchLedgers } = useLedger()

const route = useRoute()

const isActiveRoute = (url: string) => {
  if (url === '/') {
    return route.path === '/'
  }
  return route.path.startsWith(url)
}

const navMainItems = [
  { key: 'dashboard', url: '/', icon: LayoutDashboard },
  { key: 'transactions', url: '/journal', icon: Wallet },
  { key: 'accounts', url: '/accounts', icon: Calendar },
  { key: 'cards', url: '/credit-cards', icon: CreditCard },
  { key: 'budgets', url: '/budgets', icon: LineChart },
  { key: 'investments', url: '/investments', icon: Wallet }
]

const navProjectsItems = [
  { key: 'ledgerMain', url: '/ledgers', icon: Tag },
  { key: 'reports', url: '/reports', icon: LineChart }
]

const navSecondaryItems = [
  { key: 'categories', url: '/categories', icon: Tag },
  { key: 'settings', url: '/settings', icon: Settings }
]

const navMain = computed(() =>
  navMainItems.map((item) => ({
    title: t(`sidebar.main.${item.key}`),
    url: item.url,
    icon: item.icon,
    isActive: isActiveRoute(item.url)
  }))
)
const navProjects = computed(() =>
  navProjectsItems.map((item) => {
    const title =
      item.key === 'ledgerMain'
        ? currentLedger.value?.name || t('ledgers.sidebar.placeholder')
        : t(`sidebar.projects.${item.key}`)
    return {
      title,
      url: item.url,
      icon: item.icon,
      isActive: isActiveRoute(item.url)
    }
  })
)

onMounted(async () => {
  if (currentLedgerId.value && ledgers.value.length === 0) {
    await fetchLedgers()
  }
})
const navSecondary = computed(() =>
  navSecondaryItems.map((item) => ({
    title: t(`sidebar.secondary.${item.key}`),
    url: item.url,
    icon: item.icon,
    isActive: isActiveRoute(item.url)
  }))
)
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
        <SidebarGroupLabel>{{ t('sidebar.groups.projects') }}</SidebarGroupLabel>
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
              <span>{{ t('sidebar.secondary.feedback') }}</span>
            </NuxtLink>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
      <NavUser />
    </SidebarFooter>

    <SidebarRail />
  </Sidebar>
</template>
