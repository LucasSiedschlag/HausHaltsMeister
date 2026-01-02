<script setup lang="ts">
import type { Component } from 'vue'
import { SidebarGroup, SidebarGroupContent, SidebarGroupLabel, SidebarMenu, SidebarMenuButton, SidebarMenuItem } from '@shared/components/ui/sidebar'

type NavItem = {
  title: string
  url: string
  icon: Component
  isActive?: boolean
}

defineProps<{
  items: NavItem[]
}>()
</script>

<template>
  <SidebarGroup>
    <SidebarGroupLabel>Administração</SidebarGroupLabel>
    <SidebarGroupContent>
      <SidebarMenu>
        <SidebarMenuItem v-for="item in items" :key="item.url">
          <SidebarMenuButton as-child :is-active="item.isActive" :tooltip="item.title">
            <NuxtLink :to="item.url">
              <component :is="item.icon" />
              <span>{{ item.title }}</span>
            </NuxtLink>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarGroupContent>
  </SidebarGroup>
</template>
