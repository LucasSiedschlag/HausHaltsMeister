<template>
  <header class="sticky top-0 z-40 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/80">
    <div class="mx-auto flex h-14 w-full max-w-6xl items-center gap-3 px-4 md:h-16 md:px-6 lg:px-8">
      <div class="flex items-center gap-3">
        <Sheet>
          <SheetTrigger as-child>
            <Button variant="outline" size="icon" class="shrink-0 md:hidden">
              <Menu class="h-5 w-5" />
              <span class="sr-only">Toggle navigation menu</span>
            </Button>
          </SheetTrigger>
          <SheetContent side="left" class="flex w-72 flex-col gap-4 p-4 sm:w-80">
            <SheetClose as-child>
              <NuxtLink to="/" class="flex items-center gap-2 text-base font-semibold">
                <GalleryVerticalEnd class="h-5 w-5" />
                <span>HausHaltsMeister</span>
              </NuxtLink>
            </SheetClose>
            <nav class="flex flex-col gap-1 text-sm font-medium">
              <SheetClose
                v-for="item in navigation"
                :key="item.href"
                as-child
              >
                <NuxtLink
                  :to="item.href"
                  class="flex items-center gap-3 rounded-md px-3 py-2 text-muted-foreground transition-colors hover:bg-muted/60 hover:text-foreground"
                  active-class="bg-primary/10 text-primary"
                >
                  <component :is="getIcon(item.icon)" class="h-4 w-4" />
                  {{ item.title }}
                </NuxtLink>
              </SheetClose>
            </nav>
          </SheetContent>
        </Sheet>
        <NuxtLink to="/" class="flex items-center gap-2 text-base font-semibold md:hidden">
          <GalleryVerticalEnd class="h-5 w-5" />
          <span>HausHaltsMeister</span>
        </NuxtLink>
      </div>
      <div class="flex-1">
        <!-- Optional header content -->
      </div>
      <div class="flex items-center gap-2">
        <ThemeToggle />
        <DropdownMenu>
          <DropdownMenuTrigger as-child>
            <Button variant="secondary" size="icon" class="rounded-full border border-border/60">
              <CircleUser class="h-5 w-5" />
              <span class="sr-only">Toggle user menu</span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" :side-offset="8">
            <DropdownMenuLabel>My Account</DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem>Settings</DropdownMenuItem>
            <DropdownMenuItem>Support</DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem>Logout</DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { mainNavigation } from '~/layers/core/utils/navigation'
import ThemeToggle from './ThemeToggle.vue'
import * as LucideIcons from 'lucide-vue-next'
import { Menu, CircleUser, GalleryVerticalEnd } from 'lucide-vue-next'

const navigation = mainNavigation

function getIcon(name: string | undefined) {
  if (!name) return 'Circle'
  // @ts-ignore
  return LucideIcons[name] || LucideIcons.Circle
}
</script>
