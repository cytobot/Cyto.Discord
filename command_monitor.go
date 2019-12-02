package main

import (
	pb "github.com/cytobot/rpc/manager"
	"github.com/lampjaw/discordgobot"
)

type CommandMonitor struct {
	commandDefinitions map[string]*pb.CommandDefinition
	bot                *discordgobot.Gobot
	natsManager        *NatsManager
}

func NewCommandMonitor(managerClient *ManagerClient, bot *discordgobot.Gobot, natsManager *NatsManager) (*CommandMonitor, error) {
	definitions, err := managerClient.GetCommandDefinitions()
	if err != nil {
		return nil, err
	}

	monitor := &CommandMonitor{
		commandDefinitions: make(map[string]*pb.CommandDefinition, 0),
		bot:                bot,
		natsManager:        natsManager,
	}

	monitor.commandDefinitionsUpdated(definitions)

	return monitor, nil
}

func (m *CommandMonitor) commandDefinitionsUpdated(updatedDefinitions []*pb.CommandDefinition) {
	for _, updatedDef := range updatedDefinitions {
		botDefinition := m.convertToBotCommandDefinition(updatedDef)
		m.bot.UpdateCommandDefinition(botDefinition)

		m.commandDefinitions[updatedDef.CommandID] = updatedDef
	}

	for _, def := range m.commandDefinitions {
		if !containsCommandDefinition(def, updatedDefinitions) {
			m.bot.RemoveCommand(def.CommandID)

			delete(m.commandDefinitions, def.CommandID)
		}
	}
}

func containsCommandDefinition(targetDefinition *pb.CommandDefinition, sourceDefinitions []*pb.CommandDefinition) bool {
	if targetDefinition == nil || len(sourceDefinitions) == 0 {
		return false
	}

	for _, sourceDefinition := range sourceDefinitions {
		if targetDefinition.CommandID == sourceDefinition.CommandID {
			return true
		}
	}

	return false
}

func (m *CommandMonitor) convertToBotCommandDefinition(protoDef *pb.CommandDefinition) *discordgobot.CommandDefinition {
	botArgs := make([]discordgobot.CommandDefinitionArgument, 0)
	for _, paramDef := range protoDef.ParameterDefinitions {
		newArg := discordgobot.CommandDefinitionArgument{
			Optional: paramDef.Optional,
			Pattern:  paramDef.Pattern,
			Alias:    paramDef.Name,
		}
		botArgs = append(botArgs, newArg)
	}

	var permissionLevel = discordgobot.PERMISSION_USER

	switch protoDef.PermissionLevel {
	case pb.CommandDefinition_MODERATOR:
		permissionLevel = discordgobot.PERMISSION_MODERATOR
	case pb.CommandDefinition_ADMIN:
		permissionLevel = discordgobot.PERMISSION_ADMIN
	case pb.CommandDefinition_OWNER:
		permissionLevel = discordgobot.PERMISSION_OWNER
	}

	return &discordgobot.CommandDefinition{
		CommandID:               protoDef.CommandID,
		Description:             protoDef.Description,
		Triggers:                protoDef.Triggers,
		ExposureLevel:           discordgobot.EXPOSURE_EVERYWHERE,
		Unlisted:                protoDef.Unlisted,
		DisableTriggerOnMention: false,
		PermissionLevel:         permissionLevel,
		Arguments:               botArgs,
		Callback:                m.commandExecutionHandler,
	}
}

func (m *CommandMonitor) commandExecutionHandler(bot *discordgobot.Gobot, client *discordgobot.DiscordClient, payload discordgobot.CommandPayload) {
	m.natsManager.SendWorkerMessage("command", payload.CommandID, payload.Message, payload.Arguments)
}
