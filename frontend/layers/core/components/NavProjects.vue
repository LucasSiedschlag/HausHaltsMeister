<script setup lang="ts">
import type { Component } from 'vue'
import { SidebarMenu, SidebarMenuButton, SidebarMenuItem } from '@shared/components/ui/sidebar'

defineProps<{
  items: NavItem[]
}>()

type NavItem = {
  title: string
  url: string
  icon: Component
  iconClass?: string
  isActive?: boolean
}
</script>

<template>
  <SidebarMenu>
    <SidebarMenuItem v-for="item in items" :key="item.url">
      <SidebarMenuButton as-child :is-active="item.isActive" :tooltip="item.title">
        <NuxtLink :to="item.url">
          <component :is="item.icon" :class="item.iconClass" />
          <span>{{ item.title }}</span>
        </NuxtLink>
      </SidebarMenuButton>
    </SidebarMenuItem>
  </SidebarMenu>
</template>
