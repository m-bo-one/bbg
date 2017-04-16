from django.contrib import admin
from .models import BBGUser, Tank


@admin.register(BBGUser)
class BBGUserAdmin(admin.ModelAdmin):
    pass


@admin.register(Tank)
class TankAdmin(admin.ModelAdmin):

    list_display = ('name', 'tkey', 'lvl', 'kill_count', 'created_at')

    readonly_fields = ('tkey', 'lvl', 'kill_count', 'total_steps',
                       'created_at',
                       'x', 'y', 'health', 'fire_rate', 'speed', 'direction',
                       'width', 'height', 'angle',
                       'gun_damage', 'gun_bullets', 'gun_distance', 'nickname')

    fieldsets = (
        ('Tank info', {
            'fields': ('player', 'name', 'tkey', 'lvl', 'kill_count',
                       'total_steps', 'created_at'),
        }),
        ('Tank current position', {
            'classes': ('extrapretty',),
            'fields': ('nickname',
                       'x', 'y',
                       'health', 'fire_rate', 'speed', 'direction',
                       'width', 'height',
                       'angle',
                       'gun_damage', 'gun_bullets', 'gun_distance')
        }),
    )
