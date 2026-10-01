import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import BtCloseBtn from './index.vue'

describe('BtCloseBtn', () => {
  it('renders the close button', () => {
    const wrapper = mount(BtCloseBtn)
    expect(wrapper.find('.bt-close-btn').exists()).toBe(true)
    expect(wrapper.find('.close-icon').exists()).toBe(true)
  })

  it('emits click event when clicked', async () => {
    const wrapper = mount(BtCloseBtn)
    await wrapper.find('.bt-close-btn').trigger('click')
    expect(wrapper.emitted('click')).toBeTruthy()
    expect(wrapper.emitted('click')!.length).toBe(1)
  })

  it('passes event object in click emission', async () => {
    const wrapper = mount(BtCloseBtn)
    await wrapper.find('.bt-close-btn').trigger('click')
    const emitted = wrapper.emitted('click')!
    expect(emitted[0][0]).toBeInstanceOf(Event)
  })
})
