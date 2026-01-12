<script setup lang="ts">
import { computed } from 'vue'
import { Button } from '@shared/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@shared/components/ui/dropdown-menu'
import { MoreHorizontal } from 'lucide-vue-next'

export type CreditCardTileData = {
  id: string
  ledger_id: string
  parent_account_id: string
  liability_account_id: string
  label?: string | null
  brand: string
  last4?: string | null
  holder_name?: string | null
  active: boolean
  closing_day: number
  due_day: number
  color?: string | null
  style?: string | null
}

const props = withDefaults(defineProps<{
  card: CreditCardTileData
  accountName: string
  networkName?: string
  clickable?: boolean
  to?: string
  canEdit?: boolean
  showActions?: boolean
}>(), {
  clickable: true,
  canEdit: false,
  showActions: true
})

const emit = defineEmits<{
  (event: 'edit', value: CreditCardTileData): void
  (event: 'delete', value: CreditCardTileData): void
}>()

const { t } = useI18n()

const resolveNetwork = (code?: string | null) => (code || '').toLowerCase()

const networkKey = computed(() => {
  const normalized = resolveNetwork(props.card.brand)
  if (normalized.includes('master')) return 'mastercard'
  if (normalized.includes('amex') || normalized.includes('american')) return 'amex'
  if (normalized.includes('elo')) return 'elo'
  if (normalized.includes('hiper')) return 'hipercard'
  if (normalized.includes('visa')) return 'visa'
  return 'default'
})

const cardTitle = computed(() => props.card.label || props.accountName || t('creditCards.detail.unknownCard'))

const cardSubtitle = computed(() => {
  if (props.card.label) {
    return props.card.holder_name || props.accountName || t('creditCards.table.noNickname')
  }
  return props.card.holder_name || t('creditCards.table.noNickname')
})

const cardTheme = computed(() => {
  const key = networkKey.value
  if (key === 'visa') return 'from-slate-900 via-slate-800 to-slate-700'
  if (key === 'mastercard') return 'from-red-700 via-orange-600 to-amber-500'
  if (key === 'amex') return 'from-emerald-700 via-teal-600 to-cyan-500'
  if (key === 'elo') return 'from-neutral-900 via-neutral-700 to-zinc-600'
  if (key === 'hipercard') return 'from-rose-700 via-red-600 to-red-500'
  return 'from-slate-800 via-slate-700 to-slate-600'
})

const isClickable = computed(() => Boolean(props.to) && props.clickable)

const handleNavigate = () => {
  if (!isClickable.value || !props.to) return
  navigateTo(props.to)
}

const handleKeydown = (event: KeyboardEvent) => {
  if (!isClickable.value || !props.to) return
  if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault()
    navigateTo(props.to)
  }
}

const handleEdit = (event: MouseEvent) => {
  event.stopPropagation()
  emit('edit', props.card)
}

const handleDelete = (event: MouseEvent) => {
  event.stopPropagation()
  emit('delete', props.card)
}
</script>

<template>
  <div
    class="relative overflow-hidden rounded-2xl bg-gradient-to-br p-5 text-white shadow-lg"
    :class="[
      cardTheme,
      isClickable ? 'cursor-pointer transition-transform hover:-translate-y-0.5' : 'cursor-default'
    ]"
    :role="isClickable ? 'button' : undefined"
    :tabindex="isClickable ? 0 : undefined"
    @click="handleNavigate"
    @keydown="handleKeydown"
  >
    <div class="absolute -right-8 -top-8 h-24 w-24 rounded-full bg-white/10"></div>
    <div class="absolute -bottom-14 -left-10 h-32 w-32 rounded-full bg-white/10"></div>
    <div class="relative z-10 flex items-start justify-between gap-3">
      <div>
        <p class="text-[10px] uppercase tracking-[0.3em] text-white/60">
          {{ t('creditCards.table.card') }}
        </p>
        <p class="text-lg font-semibold">
          {{ cardTitle }}
        </p>
        <p class="text-xs text-white/70">
          {{ cardSubtitle }}
        </p>
      </div>
      <DropdownMenu v-if="showActions">
        <DropdownMenuTrigger as-child>
          <Button
            variant="ghost"
            size="icon"
            class="text-white/80 hover:text-white hover:bg-white/10"
            @click.stop
          >
            <MoreHorizontal class="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" @click.stop>
          <DropdownMenuItem :disabled="!canEdit" @click.stop="handleEdit">
            {{ t('creditCards.actions.edit') }}
          </DropdownMenuItem>
          <DropdownMenuItem :disabled="!canEdit" @click.stop="handleDelete">
            {{ t('creditCards.actions.delete') }}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>

    <div class="relative z-10 mt-6 flex items-center justify-between gap-3">
      <div class="flex items-center gap-2">
        <div class="flex h-8 w-12 items-center justify-center rounded-md bg-white/15">
          <svg v-if="networkKey === 'visa'" viewBox="0 0 64 24" class="h-4 w-9 fill-white" aria-hidden="true">
            <rect width="64" height="24" rx="4" fill="white" fill-opacity="0.15" />
            <text x="8" y="17" font-size="12" font-family="Arial, sans-serif" fill="white">VISA</text>
          </svg>
          <svg v-else-if="networkKey === 'mastercard'" viewBox="0 0 32 20" class="h-4 w-8" aria-hidden="true">
            <circle cx="12" cy="10" r="7" fill="#FFB347" />
            <circle cx="20" cy="10" r="7" fill="#FF5F6D" />
          </svg>
          <svg v-else-if="networkKey === 'amex'" viewBox="0 0 48 24" class="h-4 w-9" aria-hidden="true">
            <rect width="48" height="24" rx="4" fill="white" fill-opacity="0.15" />
            <text x="6" y="17" font-size="11" font-family="Arial, sans-serif" fill="white">AMEX</text>
          </svg>
          <svg v-else-if="networkKey === 'elo'" viewBox="0 0 40 20" class="h-4 w-8" aria-hidden="true">
            <circle cx="12" cy="10" r="6" fill="#00B4D8" />
            <circle cx="20" cy="10" r="6" fill="#FFC857" />
            <circle cx="28" cy="10" r="6" fill="#F94144" />
          </svg>
          <svg v-else-if="networkKey === 'hipercard'" viewBox="0 0 48 24" class="h-4 w-9" aria-hidden="true">
            <rect width="48" height="24" rx="4" fill="white" fill-opacity="0.15" />
            <text x="4" y="17" font-size="9" font-family="Arial, sans-serif" fill="white">HIPERCARD</text>
          </svg>
          <span v-else class="text-[10px] font-semibold uppercase tracking-widest text-white/80">
            {{ card.brand }}
          </span>
        </div>
        <span class="text-[10px] uppercase tracking-[0.3em] text-white/60">
          {{ networkName || card.brand }}
        </span>
      </div>
      <span class="text-xs tracking-[0.35em] text-white/80">
        .... .... .... {{ card.last4 || '----' }}
      </span>
    </div>

    <div class="relative z-10 mt-6 text-xs text-white/70">
      <div class="flex items-center justify-between">
        <div>{{ t('creditCards.table.closingDay', { day: card.closing_day }) }}</div>
        <div>{{ t('creditCards.table.dueDay', { day: card.due_day }) }}</div>
      </div>
    </div>
  </div>
</template>
