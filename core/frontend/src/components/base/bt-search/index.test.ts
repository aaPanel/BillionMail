import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import BtSearch from './index.vue'

// Stub NInput that renders slots so inner template elements are visible
const NInput = {
  template: `<div v-bind="$attrs"><slot name="prefix" /><input /><slot name="suffix" /></div>`,
  props: ['modelValue', 'value'],
  emits: ['update:modelValue', 'update:value'],
}

describe('BtSearch', () => {
  const mountComponent = (props = {}) => {
    return mount(BtSearch, {
      props,
      global: {
        stubs: {
          'n-input': NInput,
        },
      },
    })
  }

  it('renders with default width of 240px', () => {
    const wrapper = mountComponent()
    expect(wrapper.find('.bt-search').exists()).toBe(true)
    expect(wrapper.find('.bt-search').attributes('style')).toContain('240px')
  })

  it('renders with custom width', () => {
    const wrapper = mountComponent({ width: 400 })
    expect(wrapper.find('.bt-search').attributes('style')).toContain('400px')
  })

  it('renders search icon in suffix by default (prefix=false)', () => {
    const wrapper = mountComponent()
    expect(wrapper.find('.i-mdi-search').exists()).toBe(true)
  })

  it('renders search icon in prefix when prefix prop is true', () => {
    const wrapper = mountComponent({ prefix: true })
    expect(wrapper.find('.i-mdi-search').exists()).toBe(true)
  })

  it('emits search event when search icon clicked', async () => {
    const wrapper = mountComponent()
    const iconContainer = wrapper.find('.flex.items-center.cursor-pointer')
    expect(iconContainer.exists()).toBe(true)
    await iconContainer.trigger('click')
    expect(wrapper.emitted('search')).toBeTruthy()
    expect(wrapper.emitted('search')![0]).toEqual([''])
  })
})
