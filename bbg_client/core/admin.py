from django.contrib import admin
from .models import BBGUser, Tank, Stat


@admin.register(Stat)
class StatAdmin(admin.ModelAdmin):

    list_display = ('tank', 'event', 'created_at')


@admin.register(BBGUser)
class BBGUserAdmin(admin.ModelAdmin):
    pass


@admin.register(Tank)
class TankAdmin(admin.ModelAdmin):

    list_display = ('name', 'tkey', 'lvl', 'kda', 'created_at')

    readonly_fields = ('tkey', 'lvl', 'created_at', 'kda',
                       'x', 'y', 'health', 'fire_rate', 'speed', 'direction',
                       'width', 'height', 'angle',
                       'gun_damage', 'gun_bullets', 'gun_distance', 'nickname',
                       'kill_count', 'death_count', 'resurect_count')

    fieldsets = (
        ('Tank info', {
            'fields': ('player', 'name', 'tkey', 'lvl', 'created_at'),
        }),
        ('Tank stats', {
            'fields': ('kda', 'death_count', 'kill_count', 'resurect_count'),
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
