<script setup lang="ts">
import { computed } from 'vue'
import { Avatar, AvatarFallback, AvatarImage } from '@shared/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@shared/components/ui/dropdown-menu'
import { SidebarMenu, SidebarMenuButton, SidebarMenuItem } from '@shared/components/ui/sidebar'
import { ChevronsUpDown, LogOut, Settings, UserRound } from 'lucide-vue-next'
import { useAuth } from '#layers/auth/composables/useAuth'

const props = withDefaults(defineProps<{
  name?: string
  email?: string
  avatar?: string
}>(), {
  name: 'HausHaltsMeister',
  email: 'conta@exemplo.com',
  avatar: ''
})

const { user, logout } = useAuth()
const router = useRouter()

const displayName = computed(() => user.value?.display_name || props.name)
const displayEmail = computed(() => user.value?.email || props.email)
const displayAvatar = computed(() => user.value?.avatar_url || props.avatar)

const handleLogout = async () => {
  await logout()
  await router.push('/auth/login')
}
</script>

<template>
  <SidebarMenu>
    <SidebarMenuItem>
      <DropdownMenu>
        <DropdownMenuTrigger as-child>
          <SidebarMenuButton size="lg">
            <Avatar class="h-8 w-8 rounded-lg">
              <AvatarImage :src="displayAvatar" :alt="displayName" />
              <AvatarFallback class="rounded-lg bg-primary/10 text-primary">
                <UserRound class="h-4 w-4" />
              </AvatarFallback>
            </Avatar>
            <div class="grid flex-1 text-left text-sm leading-tight">
              <span class="truncate font-semibold">{{ displayName }}</span>
              <span class="truncate text-xs text-muted-foreground">{{ displayEmail }}</span>
            </div>
            <ChevronsUpDown class="ml-auto size-4 text-muted-foreground" />
          </SidebarMenuButton>
        </DropdownMenuTrigger>
        <DropdownMenuContent class="w-56" align="end" side="right">
          <DropdownMenuLabel>Conta</DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem as-child>
            <NuxtLink to="/settings" class="flex w-full items-center">
              <Settings class="mr-2 size-4" />
              Configurações
            </NuxtLink>
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem @click="handleLogout">
            <LogOut class="mr-2 size-4" />
            Sair
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </SidebarMenuItem>
  </SidebarMenu>
</template>
