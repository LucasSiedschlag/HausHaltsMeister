<script setup lang="ts">
import type { Component } from 'vue'
import { SidebarMenu, SidebarMenuButton, SidebarMenuItem } from '@shared/components/ui/sidebar'

type NavItem = {
  title: string
  url: string
  icon: Component
}

defineProps<{
  items: NavItem[]
}>()
</script>

<template>
  <SidebarMenu>
    <SidebarMenuItem v-for="item in items" :key="item.url">
      <SidebarMenuButton as-child :tooltip="item.title">
        <NuxtLink :to="item.url">
          <component :is="item.icon" />
          <span>{{ item.title }}</span>
        </NuxtLink>
      </SidebarMenuButton>
    </SidebarMenuItem>
  </SidebarMenu>
</template>
