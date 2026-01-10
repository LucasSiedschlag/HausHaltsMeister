# Implementação de Ledger Padrão e Gestão de Membros

**Data**: 2026-01-05
**Status**: Implementado
**Objetivo**: Permitir que usuários definam um ledger padrão nas preferências e gerenciem membros/permissões

---

## Resumo Executivo

Implementar funcionalidade para:

1. Usuário selecionar um ledger padrão que será automaticamente ativado no login
2. Listar ledgers com seus membros e permissões na aba "Ledger" das configurações
3. Gerenciar acesso de outros usuários aos ledgers (convidar, editar permissão, remover)

---

## Mudanças Já Realizadas

- [x] Removido `ensureLedger()` de `frontend/app/plugins/auth-bootstrap.client.ts`

---

## Proposta de Implementação

### 1. Backend - Preferências de Usuário

#### `internal/domain/preferences/entity.go`

Adicionar campo `default_ledger_id`:

```go
type UserPreferences struct {
    // ... campos existentes
    DefaultLedgerID *string `json:"default_ledger_id"` // nullable
}
```

#### Na Migration 002

```sql
ALTER TABLE user_preferences
ADD COLUMN default_ledger_id UUID REFERENCES ledgers(id);
```

#### `internal/adapters/http/handlers/preferences.go`

Atualizar endpoints GET/PUT para incluir `default_ledger_id`.

---

### 2. Backend - API de Membros do Ledger

#### `internal/domain/ledger/models.go`

```go
type Member struct {
    LedgerID    string     `json:"ledger_id"`
    UserID      string     `json:"user_id"`
    Role        string     `json:"role"` // owner, editor, viewer
    DisplayName string     `json:"display_name"`
    Email       string     `json:"email"`
    AvatarURL   *string    `json:"avatar_url,omitempty"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}
```

#### `internal/adapters/http/handlers/ledger.go`

Novos endpoints:

| Método | Rota                             | Descrição                  |
| ------ | -------------------------------- | -------------------------- |
| GET    | `/ledgers/:ledgerId/members`     | Lista membros              |
| POST   | `/ledgers/:ledgerId/members`     | Convida membro (por email) |
| PATCH  | `/ledgers/:ledgerId/members/:id` | Atualiza permissão         |
| DELETE | `/ledgers/:ledgerId/members/:id` | Remove membro              |

---

### 3. Frontend - Composables

#### `layers/shared/composables/usePreferences.ts`

Adicionar tipo e lógica para `default_ledger_id`:

```typescript
export type UserPreferences = {
  // ... existentes
  default_ledger_id: string | null;
};
```

#### `layers/ledgers/composables/useLedgerMembers.ts` (novo)

```typescript
export const useLedgerMembers = () => {
  const members = ref<LedgerMember[]>([])
  const loading = ref(false)

  const fetchMembers = async (ledgerId: string) => { ... }
  const inviteMember = async (ledgerId: string, email: string, role: LedgerRole) => { ... }
  const updateRole = async (ledgerId: string, memberId: string, role: LedgerRole) => { ... }
  const removeMember = async (ledgerId: string, memberId: string) => { ... }

  return { members, loading, fetchMembers, inviteMember, updateRole, removeMember }
}
```

---

### 4. Frontend - Settings Page

#### `layers/core/pages/settings.vue`

Na aba "Ledger":

1. **Select de Ledger Padrão**
   - Dropdown com ledgers do usuário
   - Auto-save ao selecionar
2. **Lista de Ledgers com Membros**

   - Cards expandíveis para cada ledger
   - Mostra role do usuário atual
   - Lista de membros com avatares

3. **Modal de Convidar Membro**
   - Input de email
   - Select de role (editor/viewer)
   - Botão convidar

---

### 5. Frontend - Bootstrap de Ledger

#### `app/plugins/auth-bootstrap.client.ts`

Após login, verificar `default_ledger_id` das preferências:

```typescript
// Após carregar preferências
if (preferences.default_ledger_id) {
  await ledgerContext.setActiveLedger(preferences.default_ledger_id);
}
```

---

## Fluxo de Usuário

```mermaid
flowchart TD
    A[Login] --> B{Tem default_ledger_id?}
    B -->|Sim| C[Ativa ledger padrão]
    B -->|Não| D[Mostra LedgerSwitcher]
    C --> E[Dashboard]
    D --> F[Usuário seleciona ledger]
    F --> E
```

---

## Tarefas Ordenadas

1. [x] Migration: adicionar `default_ledger_id` na tabela
2. [x] Backend: atualizar entity e repository de preferences
3. [x] Backend: atualizar handlers de preferences
4. [x] Backend: atualizar entity e repository de ledger_members
5. [x] Backend: atualizar handlers de ledger_members
6. [x] Frontend: atualizar types em usePreferences
7. [x] Frontend: criar composable useLedgerMembers
8. [x] Frontend: implementar UI de ledger padrão em settings
9. [x] Frontend: implementar UI de lista de membros
10. [x] Frontend: implementar modal de convite
11. [x] Frontend: atualizar auth-bootstrap para usar default_ledger_id
12. [x] Testes e ajustes finais

---

## Estimativa

| Componente | Tempo          |
| ---------- | -------------- |
| Backend    | ~2-3 horas     |
| Frontend   | ~3-4 horas     |
| **Total**  | **~5-7 horas** |

---

## Verificação

- [x] Ao fazer login, o ledger padrão é ativado automaticamente
- [x] Se não há ledger padrão, o LedgerSwitcher funciona normalmente
- [x] Usuário pode alterar ledger padrão nas configurações
- [x] Lista de membros aparece corretamente para cada ledger
- [x] Convite por email funciona (owner/editor apenas)
- [x] Permissões são respeitadas (viewer não pode editar nada)
