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
import { Calendar, CreditCard, Crown, Eye, LayoutDashboard, LineChart, Pencil, Send, Settings, Tag, Wallet } from 'lucide-vue-next'
import { useRoute } from '#imports'
import NavMain from './NavMain.vue'
import NavProjects from './NavProjects.vue'
import NavSecondary from './NavSecondary.vue'
import NavUser from './NavUser.vue'
import { useLedger } from '@shared/composables/useLedger'

const { t } = useI18n()
const { currentLedger, currentLedgerId, ledgers, fetchLedgers } = useLedger()

const ledgerName = computed(() => currentLedger.value?.name || t('ledgers.sidebar.placeholder'))
const roleMeta = computed(() => {
  const role = currentLedger.value?.role
  if (role === 'owner') {
    return { icon: Crown, class: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300' }
  }
  if (role === 'editor') {
    return { icon: Pencil, class: 'border-sky-500/30 bg-sky-500/10 text-sky-700 dark:text-sky-300' }
  }
  if (role === 'viewer') {
    return { icon: Eye, class: 'border-muted-foreground/30 bg-muted/50 text-muted-foreground' }
  }
  return { icon: Eye, class: 'border-muted-foreground/30 bg-muted/50 text-muted-foreground' }
})
const roleIconClass = computed(() => {
  const role = currentLedger.value?.role
  if (role === 'owner') {
    return 'text-emerald-600 dark:text-emerald-300'
  }
  if (role === 'editor') {
    return 'text-sky-600 dark:text-sky-300'
  }
  if (role === 'viewer') {
    return 'text-muted-foreground'
  }
  return 'text-muted-foreground'
})

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
    const icon = item.key === 'ledgerMain' && currentLedger.value ? roleMeta.value.icon : item.icon
    const iconClass =
      item.key === 'ledgerMain' && currentLedger.value ? roleIconClass.value : undefined
    return {
      title,
      url: item.url,
      icon,
      iconClass,
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
      <div
        data-sidebar="footer"
        class="rounded-md border border-border/60 bg-muted/40 px-3 py-2 text-xs text-muted-foreground"
      >
        <p class="flex items-center gap-2 font-medium text-foreground">
          <span class="inline-flex h-5 w-5 items-center justify-center rounded-full border" :class="roleMeta.class">
            <component :is="roleMeta.icon" class="h-3.5 w-3.5" />
          </span>
          <span class="truncate">{{ ledgerName }}</span>
        </p>
      </div>
      <NavUser />
    </SidebarFooter>

    <SidebarRail />
  </Sidebar>
</template>
