<script setup lang="ts">
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '~/layers/shared/components/ui/dialog'
import { Button } from '~/layers/shared/components/ui/button'
import type { PicuinhaEntry } from '../types/picuinha'

interface Props {
  open: boolean
  entry?: PicuinhaEntry | null
  submitting?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  entry: null,
  submitting: false,
})

const emit = defineEmits<{
  'update:open': [value: boolean]
  confirm: []
}>()
</script>

<template>
  <Dialog :open="props.open" @update:open="emit('update:open', $event)">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Excluir lançamento</DialogTitle>
        <DialogDescription>
          Tem certeza que deseja excluir este lançamento? Esta ação não pode ser desfeita.
        </DialogDescription>
      </DialogHeader>
      <DialogFooter class="mt-6">
        <Button variant="outline" :disabled="props.submitting" @click="emit('update:open', false)">
          Cancelar
        </Button>
        <Button variant="destructive" :disabled="props.submitting" @click="emit('confirm')">
          Excluir
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
