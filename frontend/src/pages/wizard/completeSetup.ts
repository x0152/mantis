import { api } from '@/api'
import type { State } from './types'
import { SEED_PLANS, SEED_SKILLS } from './seeds'
import { makeUniqueID, normalizeBaseUrl, pollConnections, suggestEndpointId } from './utils'

export async function completeSetup(state: State): Promise<void> {
  const provider = state.provider
  const baseUrl = (provider === 'openai' ? state.openaiBaseUrl : state.gonkaNodeUrl).trim()
  const apiKey = (provider === 'openai' ? state.openaiApiKey : state.gonkaPrivateKey).trim()

  const existingEndpoints = await api.llmConnections.list()
  const existingByURL = existingEndpoints.find(e => normalizeBaseUrl(e.baseUrl) === normalizeBaseUrl(baseUrl))
  const endpointID = existingByURL
    ? existingByURL.id
    : makeUniqueID(suggestEndpointId(baseUrl, provider), new Set(existingEndpoints.map(e => e.id)))

  if (existingByURL) {
    await api.llmConnections.update(endpointID, { provider, baseUrl, apiKey })
  } else {
    await api.llmConnections.create({ id: endpointID, provider, baseUrl, apiKey })
  }

  const existingModels = await api.models.list()
  const findOrCreate = async (name: string) => {
    const found = existingModels.find(m => m.connectionId === endpointID && m.name === name)
    return (
      found ??
      (await api.models.create({
        connectionId: endpointID,
        name,
        thinkingMode: '',
        contextWindow: 0,
        reserveTokens: 0,
        compactTokens: 0,
      }))
    )
  }

  const validRows = state.modelRows.filter(r => r.name.trim())
  const created: Record<string, { id: string }> = {}
  for (const row of validRows) {
    created[row.name.trim()] = await findOrCreate(row.name.trim())
  }

  const chatRow = validRows.find(r => r.role === 'chat')
  if (!chatRow) throw new Error('Pick a chat model first.')
  const summaryRow = validRows.find(r => r.role === 'summary')
  const visionRow = validRows.find(r => r.role === 'vision')
  const primaryModelId = created[chatRow.name.trim()].id
  const summaryModelId = summaryRow ? created[summaryRow.name.trim()].id : ''
  const imageModelId = visionRow ? created[visionRow.name.trim()].id : ''

  const profileName = `Main profile (${endpointID})`
  const existingPresets = await api.presets.list()
  const existingPreset =
    existingPresets.find(p => p.name === profileName) ?? existingPresets.find(p => p.name === 'Default')
  const preset = existingPreset
    ? await api.presets.update(existingPreset.id, {
        name: profileName,
        chatModelId: primaryModelId,
        summaryModelId,
        imageModelId,
        fallbackModelId: '',
        temperature: null,
        systemPrompt: '',
      })
    : await api.presets.create({
        name: profileName,
        chatModelId: primaryModelId,
        summaryModelId,
        imageModelId,
        fallbackModelId: '',
        temperature: null,
        systemPrompt: '',
      })

  const connList = await pollConnections(['base', 'browser'], 15_000).catch(
    () => [] as Awaited<ReturnType<typeof api.connections.list>>,
  )
  const connByName = new Map(connList.map(c => [c.name, c]))
  for (const profileId of Object.keys(SEED_SKILLS)) {
    const conn = connByName.get(profileId)
    if (!conn) continue
    for (const sk of SEED_SKILLS[profileId]) {
      await api.skills.create({ ...sk, connectionId: conn.id }).catch(() => {})
    }
  }

  for (const plan of SEED_PLANS) {
    await api.plans.create(plan).catch(() => {})
  }

  await api.channels.update('chat', { name: 'Chat', token: '', allowedUserIds: [] })

  const token = state.tgToken.trim()
  if (token && !state.tgSkip) {
    const envIds = state.tgEnvUserIds
      .split(',')
      .map(s => parseInt(s.trim(), 10))
      .filter(n => !isNaN(n))
    const ids = state.tgLinkedUser ? [state.tgLinkedUser.id, ...envIds] : envIds
    const uniqueIds = Array.from(new Set(ids))
    await api.channels
      .create({ type: 'telegram', name: 'Telegram', token, modelId: '', presetId: '', allowedUserIds: uniqueIds })
      .catch(() => {})
  }

  await api.settings.update({
    chatPresetId: preset.id,
    serverPresetId: preset.id,
    memoryEnabled: true,
    userMemories: [],
  })
}
